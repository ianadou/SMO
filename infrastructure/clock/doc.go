// Package clock provides production implementations of the
// domain ports.Clock interface.
//
// The default implementation simply wraps time.Now(). Tests should
// inject their own deterministic clock instead of using this one.
package clock
