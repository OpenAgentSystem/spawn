// Package supplylayer contains OpenAgentSystem policy gates for this hard fork.
//
// It is intentionally isolated from the upstream runtime adapters in the first
// landing commit: callers can evaluate launch requests before invoking a spwn
// runtime, while the fork keeps upstream merge pressure low.
package supplylayer
