package mappers

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// optionalTimeToPg converts an optional *time.Time into pgtype.Timestamptz.
// Nil becomes the NULL representation.
func optionalTimeToPg(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// optionalTimeFromPg converts a nullable pgtype.Timestamptz into an
// optional *time.Time. NULL becomes nil.
func optionalTimeFromPg(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}
