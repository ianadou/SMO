// Package ratelimit provides a Gin middleware that enforces per-IP
// fixed-window rate limits on selected routes, backed by Redis.
//
// The middleware is declarative: a Config maps Gin route patterns
// (e.g., "/api/v1/auth/login") to a (limit, window) pair. Routes not
// listed in the config pass through unchanged.
//
// Redis is a soft dependency (see ADR 0002): if Redis errors at
// runtime, the middleware logs a throttled WARN and allows the
// request through. Rate limiting is a usage policy, not a security
// boundary; the primary defenses (password hashing, account lockout,
// JWT auth) remain in effect when the cache layer hiccups.
package ratelimit
