// Package postgres provides the PostgreSQL infrastructure for the SMO
// application: connection pool setup with startup retries, embedded
// migrations applied via goose, and the wiring point for repositories
// that use the sqlc-generated queries.
//
// Sub-packages:
//   - generated/    → sqlc-generated type-safe query code (do not edit)
//   - mappers/      → translation between sqlc structs and domain entities
//   - repositories/ → concrete implementations of the domain repository ports
package postgres
