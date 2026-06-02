package runtimes

import (
	"errors"
	"fmt"
	"strings"
)

const AdapterSpecV1 = "oas.agent_adapter.v1"

type RuntimeSurface string

const (
	RuntimeSurfaceCLI          RuntimeSurface = "cli"
	RuntimeSurfaceIDEExtension RuntimeSurface = "ide_extension"
	RuntimeSurfaceLocalServer  RuntimeSurface = "local_server"
	RuntimeSurfaceMixed        RuntimeSurface = "mixed"
)

type ToSReviewState string

const (
	ToSReviewAllowed     ToSReviewState = "allowed"
	ToSReviewConditional ToSReviewState = "conditional"
	ToSReviewBlocked     ToSReviewState = "blocked"
	ToSReviewUnknown     ToSReviewState = "unknown"
)

type IntendedUse string

const (
	IntendedUseInternalRnD      IntendedUse = "internal_r_and_d"
	IntendedUseControlledPoC    IntendedUse = "controlled_poc"
	IntendedUseCustomerDelivery IntendedUse = "customer_delivery"
	IntendedUseSaaSEmbedded     IntendedUse = "saas_embedded"
)

type AdapterSpec struct {
	SchemaVersion  string
	AgentFamily    string
	AdapterVersion string
	RuntimeSurface RuntimeSurface
	LegalEvidence  LegalEvidence
}

type LegalEvidence struct {
	ToSURL                  string
	PrivacyURL              string
	LicenseURL              string
	ReviewState             ToSReviewState
	IntendedUses            []IntendedUse
	CustomerDataAllowed     bool
	SourceCodeAllowed       bool
	ProductionLogAllowed    bool
	TrainingOptOutAvailable bool
}

type DispatchPolicy struct {
	CustomerData  bool
	SourceCode    bool
	ProductionLog bool
	IntendedUse   IntendedUse
}

var (
	ErrAdapterSpecInvalid = errors.New("adapter spec invalid")
	ErrToSGateDenied      = errors.New("tos gate denied")
)

func ValidateAdapterSpec(a Adapter) error {
	spec := a.Spec
	missing := []string{}
	if strings.TrimSpace(a.Name) == "" {
		missing = append(missing, "name")
	}
	if spec.SchemaVersion != AdapterSpecV1 {
		missing = append(missing, "adapter.schema_version")
	}
	if strings.TrimSpace(spec.AgentFamily) == "" {
		missing = append(missing, "adapter.agent_family")
	}
	if strings.TrimSpace(spec.AdapterVersion) == "" {
		missing = append(missing, "adapter.adapter_version")
	}
	if !validRuntimeSurface(spec.RuntimeSurface) {
		missing = append(missing, "adapter.runtime_surface")
	}
	if strings.TrimSpace(spec.LegalEvidence.ToSURL) == "" {
		missing = append(missing, "adapter.legal_evidence.tos_url")
	}
	if strings.TrimSpace(spec.LegalEvidence.PrivacyURL) == "" {
		missing = append(missing, "adapter.legal_evidence.privacy_url")
	}
	if !validToSReviewState(spec.LegalEvidence.ReviewState) {
		missing = append(missing, "adapter.legal_evidence.review_state")
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w for %q: missing or invalid %s", ErrAdapterSpecInvalid, a.Name, strings.Join(missing, ", "))
	}
	if spec.AgentFamily != a.Name {
		return fmt.Errorf("%w for %q: adapter.agent_family %q must match adapter name", ErrAdapterSpecInvalid, a.Name, spec.AgentFamily)
	}
	return nil
}

func AuthorizeDispatch(adapterName string, policy DispatchPolicy) error {
	a, ok := Get(adapterName)
	if !ok {
		return fmt.Errorf("%w: unknown adapter %q", ErrToSGateDenied, adapterName)
	}
	return AuthorizeAdapterDispatch(a, policy)
}

func AuthorizeAdapterDispatch(a Adapter, policy DispatchPolicy) error {
	if err := ValidateAdapterSpec(a); err != nil {
		return err
	}
	legal := a.Spec.LegalEvidence
	switch legal.ReviewState {
	case ToSReviewAllowed:
	case ToSReviewConditional:
		if policy.CustomerData {
			return fmt.Errorf("%w: %s ToS state is conditional and customer data is not allowed", ErrToSGateDenied, a.Name)
		}
	case ToSReviewBlocked:
		return fmt.Errorf("%w: %s ToS state is blocked", ErrToSGateDenied, a.Name)
	default:
		return fmt.Errorf("%w: %s ToS state is unknown", ErrToSGateDenied, a.Name)
	}
	if policy.CustomerData && !legal.CustomerDataAllowed {
		return fmt.Errorf("%w: %s legal evidence does not allow customer data", ErrToSGateDenied, a.Name)
	}
	if policy.SourceCode && !legal.SourceCodeAllowed {
		return fmt.Errorf("%w: %s legal evidence does not allow source code", ErrToSGateDenied, a.Name)
	}
	if policy.ProductionLog && !legal.ProductionLogAllowed {
		return fmt.Errorf("%w: %s legal evidence does not allow production logs", ErrToSGateDenied, a.Name)
	}
	if policy.IntendedUse != "" && !intendedUseAllowed(policy.IntendedUse, legal.IntendedUses) {
		return fmt.Errorf("%w: %s intended use %q is not allowed by legal evidence", ErrToSGateDenied, a.Name, policy.IntendedUse)
	}
	return nil
}

func validRuntimeSurface(surface RuntimeSurface) bool {
	switch surface {
	case RuntimeSurfaceCLI, RuntimeSurfaceIDEExtension, RuntimeSurfaceLocalServer, RuntimeSurfaceMixed:
		return true
	default:
		return false
	}
}

func validToSReviewState(state ToSReviewState) bool {
	switch state {
	case ToSReviewAllowed, ToSReviewConditional, ToSReviewBlocked, ToSReviewUnknown:
		return true
	default:
		return false
	}
}

func intendedUseAllowed(want IntendedUse, allowed []IntendedUse) bool {
	for _, use := range allowed {
		if use == want {
			return true
		}
	}
	return false
}
