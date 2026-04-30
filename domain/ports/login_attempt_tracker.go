package ports

import "context"

// LoginAttemptTracker records authentication attempts and exposes the
// "is this account currently locked out?" query consumed by
// LoginOrganizerUseCase.
//
// The contract:
//
//   - IsLocked is called BEFORE the password check. If it returns true,
//     the use case short-circuits with ErrAccountLocked without ever
//     hitting the password hasher.
//   - RecordFailure is called after a credential mismatch. The
//     implementation decides when accumulated failures cross into a
//     lockout (the policy lives in the adapter's Config, not here).
//   - RecordSuccess is called after a successful authentication. The
//     implementation must clear any pending failure counter and any
//     active lockout for that email.
//
// All three methods take a normalized email (lowercase) so callers do
// not need to coordinate the casing rule with the adapter.
//
// Errors returned by IsLocked, RecordFailure or RecordSuccess MUST NOT
// fail the login flow. Adapters degrade fail-open on backend errors
// (Redis down, network blip): the WARN log is emitted by the adapter,
// the use case treats the call as a no-op and continues. Locking out
// on infrastructure trouble would turn a Redis hiccup into a global
// auth outage — see ADR 0007.
type LoginAttemptTracker interface {
	IsLocked(ctx context.Context, email string) (bool, error)
	RecordFailure(ctx context.Context, email string) error
	RecordSuccess(ctx context.Context, email string) error
}
