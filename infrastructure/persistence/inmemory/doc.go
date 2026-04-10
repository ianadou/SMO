// Package inmemory provides in-memory implementations of the domain
// repository ports.
//
// These implementations are intended for local development and tests.
// They are NOT meant to be used in production: data is lost when the
// process exits, and there is no transactional support beyond the
// simple mutex-based locking.
//
// In production, the composition root should wire the PostgreSQL
// implementations from infrastructure/persistence/postgres/repositories
// instead.
package inmemory
