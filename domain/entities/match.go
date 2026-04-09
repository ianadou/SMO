package entities

import (
	"strings"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

const (
	maxMatchTitleLength = 100
	maxVenueLength      = 200
)

// MatchID is the unique identifier of a Match.
type MatchID string

// Match represents a scheduled sports match within a group.
//
// A match has a lifecycle (see MatchStatus) that controls which operations
// are allowed at each step. The state machine logic itself is not in this
// PR; it will be added in a dedicated PR with its own tests.
type Match struct {
	id          MatchID
	groupID     GroupID
	title       string
	venue       string
	scheduledAt time.Time
	status      MatchStatus
	createdAt   time.Time
}

// NewMatch builds a Match after validating its inputs.
//
// A new match always starts in MatchStatusDraft regardless of what
// the caller passes; the status parameter is intended for rehydration
// from persistence, not for new matches. The behavior is identical here
// because we currently accept any valid status, but a future refactor
// might split NewMatch and RehydrateMatch into two separate constructors.
func NewMatch(
	id MatchID,
	groupID GroupID,
	title string,
	venue string,
	scheduledAt time.Time,
	status MatchStatus,
	createdAt time.Time,
) (*Match, error) {
	if id == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if groupID == "" {
		return nil, domainerrors.ErrInvalidID
	}

	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" || len(trimmedTitle) > maxMatchTitleLength {
		return nil, domainerrors.ErrInvalidName
	}

	trimmedVenue := strings.TrimSpace(venue)
	if trimmedVenue == "" || len(trimmedVenue) > maxVenueLength {
		return nil, domainerrors.ErrInvalidName
	}

	if scheduledAt.IsZero() {
		return nil, domainerrors.ErrInvalidDate
	}

	// Validate the status by re-parsing it. This protects against callers
	// passing a string-cast invalid status (e.g., MatchStatus("typo")).
	if _, err := ParseMatchStatus(string(status)); err != nil {
		return nil, err
	}

	if createdAt.IsZero() {
		return nil, domainerrors.ErrInvalidDate
	}

	return &Match{
		id:          id,
		groupID:     groupID,
		title:       trimmedTitle,
		venue:       trimmedVenue,
		scheduledAt: scheduledAt,
		status:      status,
		createdAt:   createdAt,
	}, nil
}

// ID returns the match identifier.
func (m *Match) ID() MatchID { return m.id }

// GroupID returns the identifier of the group this match belongs to.
func (m *Match) GroupID() GroupID { return m.groupID }

// Title returns the match title.
func (m *Match) Title() string { return m.title }

// Venue returns the venue where the match is scheduled.
func (m *Match) Venue() string { return m.venue }

// ScheduledAt returns the date and time the match is scheduled to start.
func (m *Match) ScheduledAt() time.Time { return m.scheduledAt }

// Status returns the current status of the match.
func (m *Match) Status() MatchStatus { return m.status }

// CreatedAt returns the creation timestamp of the match.
func (m *Match) CreatedAt() time.Time { return m.createdAt }
