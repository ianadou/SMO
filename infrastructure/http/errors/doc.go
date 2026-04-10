// Package errors translates domain errors into HTTP status codes and
// safe public messages.
//
// Domain errors carry rich context that is useful server-side (and
// preserved in logs) but must never be returned verbatim to clients:
// they would leak internal structure, technology choices, and
// potentially sensitive details. This package centralizes the mapping
// so every handler returns consistent responses.
package errors
