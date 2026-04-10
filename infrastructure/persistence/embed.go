// Package persistence groups the cross-driver persistence concerns of
// the SMO application: embedded migrations and anything else that is
// conceptually shared between specific database drivers.
//
// Driver-specific code (sqlc-generated queries, pgx-based repository
// implementations) lives in sub-packages such as postgres/.
package persistence

import "embed"

// MigrationsFS is the embedded filesystem containing the goose SQL
// migrations. Embedding makes the binary fully self-contained: there
// is no need to ship the migrations/ directory alongside the binary
// at deploy time.
//
// The migrations are kept in a driver-neutral location because the
// same SQL could in principle be applied by any PostgreSQL-compatible
// driver. The postgres/migrator.go file is what actually runs them
// through goose against a live pgx connection.
//
//go:embed all:migrations/*.sql
var MigrationsFS embed.FS
