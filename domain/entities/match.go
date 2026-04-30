package entities

import (
	"strings"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

const (
	maxMatchTitleLength = 100
	maxVenueLength      = 200

	// MaxParticipantsPerMatch caps the number of confirmed invitations
	// (used_at IS NOT NULL) that a single match can accumulate. AcceptInvitation
	// returns ErrMatchFull on the (N+1)th attempt; this is the FCFS policy
	// agreed in ADR 0008.
	MaxParticipantsPerMatch = 10
)

// MatchID is the unique identifier of a Match.
type MatchID string

// Match represents a scheduled sports match within a group.
//
// A match has a lifecycle (see MatchStatus) that controls which operations
// are allowed at each step. The state machine transition methods live in
// match_transitions.go.
type Match struct {
	id          MatchID
	groupID     GroupID
	title       string
	venue       string
	scheduledAt time.Time
	status      MatchStatus
	mvpPlayerID *PlayerID
	createdAt   time.Time
}

// MatchSnapshot is the full state of a Match as persisted. Used by
// RehydrateMatch to rebuild an entity from a database row without
// re-validating against the state machine — the row was valid when it
// was written, and lifecycle invariants are enforced at write time.
type MatchSnapshot struct {
	ID          MatchID
	GroupID     GroupID
	Title       string
	Venue       string
	ScheduledAt time.Time
	Status      MatchStatus
	MVPPlayerID *PlayerID
	CreatedAt   time.Time
}

// NewMatch builds a brand new Match in MatchStatusDraft. Use this from
// the create-match use case; for rehydration from persistence, call
// RehydrateMatch with a snapshot instead.
func NewMatch(
	id MatchID,
	groupID GroupID,
	title string,
	venue string,
	scheduledAt time.Time,
	createdAt time.Time,
) (*Match, error) {
	return RehydrateMatch(MatchSnapshot{
		ID:          id,
		GroupID:     groupID,
		Title:       title,
		Venue:       venue,
		ScheduledAt: scheduledAt,
		Status:      MatchStatusDraft,
		MVPPlayerID: nil,
		CreatedAt:   createdAt,
	})
}

// RehydrateMatch rebuilds a Match from its persisted snapshot. Inputs
// are validated for shape (non-empty IDs, sane lengths, valid status
// string) but not against state-machine invariants: a stored row was
// already validated when it was first written.
func RehydrateMatch(s MatchSnapshot) (*Match, error) {
	if s.ID == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if s.GroupID == "" {
		return nil, domainerrors.ErrInvalidID
	}

	trimmedTitle := strings.TrimSpace(s.Title)
	if trimmedTitle == "" || len(trimmedTitle) > maxMatchTitleLength {
		return nil, domainerrors.ErrInvalidName
	}

	trimmedVenue := strings.TrimSpace(s.Venue)
	if trimmedVenue == "" || len(trimmedVenue) > maxVenueLength {
		return nil, domainerrors.ErrInvalidName
	}

	if s.ScheduledAt.IsZero() {
		return nil, domainerrors.ErrInvalidDate
	}

	// Validate the status by re-parsing it. This protects against callers
	// passing a string-cast invalid status (e.g., MatchStatus("typo")).
	if _, err := ParseMatchStatus(string(s.Status)); err != nil {
		return nil, err
	}

	if s.CreatedAt.IsZero() {
		return nil, domainerrors.ErrInvalidDate
	}

	return &Match{
		id:          s.ID,
		groupID:     s.GroupID,
		title:       trimmedTitle,
		venue:       trimmedVenue,
		scheduledAt: s.ScheduledAt,
		status:      s.Status,
		mvpPlayerID: s.MVPPlayerID,
		createdAt:   s.CreatedAt,
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

// MVP returns the identifier of the player elected MVP for this match,
// or nil if no MVP was elected (match not yet finalized, or finalized
// with no votes).
func (m *Match) MVP() *PlayerID { return m.mvpPlayerID }

// CreatedAt returns the creation timestamp of the match.
func (m *Match) CreatedAt() time.Time { return m.createdAt }
