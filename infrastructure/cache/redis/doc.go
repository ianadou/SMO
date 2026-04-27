// Package redis provides Redis-backed cache adapters for the SMO
// repositories. Caching is wired by composition: Caching*Repository
// types implement the same domain ports as the underlying Postgres
// repositories and wrap them transparently. See ADR 0002 for the
// design and the three configuration states (disabled, enabled,
// runtime-degraded).
package redis
