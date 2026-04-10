package postgres

import (
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// RunMigrations applies all pending goose migrations to the database
// pointed at by the given pool, using the migrations found in fsys.
//
// Goose requires a *sql.DB (the standard library interface), not a
// pgxpool.Pool directly. We use the stdlib adapter from pgx to bridge
// the two without opening a second connection pool.
//
// The migrations directory inside fsys is fixed to "migrations" to
// match the embed.FS layout in package persistence.
func RunMigrations(pool *pgxpool.Pool, fsys fs.FS) error {
	db := stdlib.OpenDBFromPool(pool)
	defer func() {
		// Closing the stdlib wrapper does not close the pool itself,
		// which is what we want: the pool stays alive for the rest of
		// the application.
		_ = db.Close()
	}()

	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("postgres migrator: set dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("postgres migrator: apply migrations: %w", err)
	}

	return nil
}
