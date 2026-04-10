// Package repositories contains the PostgreSQL implementations of the
// domain repository ports.
//
// Each repository wraps the sqlc-generated Queries struct and uses the
// mappers package to translate between sqlc structs and domain entities.
// This is where the boundary between the database and the domain is
// concretely enforced.
//
// All repositories accept a generated.DBTX at construction time, which
// can be a connection, a connection pool, or a transaction. This allows
// the same repository to be used both for standalone operations and
// inside a transaction managed by a use case.
package repositories
