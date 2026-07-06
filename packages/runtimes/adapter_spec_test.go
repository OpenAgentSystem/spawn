package runtimes_test

import (
	"errors"
	"testing"

	"spwn.sh/packages/runtimes"

	_ "spwn.sh/packages/runtimes/defaults"
)

func TestBuiltInAdaptersDeclareSpecV1(t *testing.T) {
	for _, a := range runtimes.All() {
		t.Run(a.Name, func(t *testing.T) {
			if err := runtimes.ValidateAdapterSpec(a); err != nil {
				t.Fatalf("ValidateAdapterSpec: %v", err)
			}
		})
	}
}

func TestAdapterSpecValidationFailsClosedForMissingFields(t *testing.T) {
	err := runtimes.ValidateAdapterSpec(runtimes.Adapter{
		Name: "incomplete",
		Spec: runtimes.AdapterSpec{
			SchemaVersion: runtimes.AdapterSpecV1,
			AgentFamily:   "incomplete",
		},
	})
	if !errors.Is(err, runtimes.ErrAdapterSpecInvalid) {
		t.Fatalf("ValidateAdapterSpec error = %v, want ErrAdapterSpecInvalid", err)
	}
}

func TestToSGateAllowsConditionalInternalPoC(t *testing.T) {
	a := testAdapter("wave1", runtimes.ToSReviewConditional)
	err := runtimes.AuthorizeAdapterDispatch(a, runtimes.DispatchPolicy{
		IntendedUse: runtimes.IntendedUseControlledPoC,
	})
	if err != nil {
		t.Fatalf("AuthorizeAdapterDispatch: %v", err)
	}
}

func TestToSGateDeniesConditionalCustomerData(t *testing.T) {
	a := testAdapter("wave1", runtimes.ToSReviewConditional)
	err := runtimes.AuthorizeAdapterDispatch(a, runtimes.DispatchPolicy{
		CustomerData: true,
		IntendedUse:  runtimes.IntendedUseControlledPoC,
	})
	if !errors.Is(err, runtimes.ErrToSGateDenied) {
		t.Fatalf("AuthorizeAdapterDispatch error = %v, want ErrToSGateDenied", err)
	}
}

func TestToSGateDeniesUnknownAndBlockedStates(t *testing.T) {
	for _, state := range []runtimes.ToSReviewState{
		runtimes.ToSReviewUnknown,
		runtimes.ToSReviewBlocked,
	} {
		t.Run(string(state), func(t *testing.T) {
			err := runtimes.AuthorizeAdapterDispatch(testAdapter("wave1-"+string(state), state), runtimes.DispatchPolicy{})
			if !errors.Is(err, runtimes.ErrToSGateDenied) {
				t.Fatalf("AuthorizeAdapterDispatch error = %v, want ErrToSGateDenied", err)
			}
		})
	}
}

func testAdapter(name string, state runtimes.ToSReviewState) runtimes.Adapter {
	return runtimes.Adapter{
		Name: name,
		Spec: runtimes.AdapterSpec{
			SchemaVersion:  runtimes.AdapterSpecV1,
			AgentFamily:    name,
			AdapterVersion: "1.0.0",
			RuntimeSurface: runtimes.RuntimeSurfaceCLI,
			LegalEvidence: runtimes.LegalEvidence{
				ToSURL:       "https://example.com/terms",
				PrivacyURL:   "https://example.com/privacy",
				ReviewState:  state,
				IntendedUses: []runtimes.IntendedUse{runtimes.IntendedUseControlledPoC},
			},
		},
	}
}
