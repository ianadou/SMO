package entities

import domainerrors "github.com/ianadou/smo/domain/errors"

// MatchStatus represents the lifecycle stage of a match.
//
// Valid transitions: draft -> open -> teams_ready -> in_progress -> completed -> closed.
//
// Transition logic itself lives on the Match entity in a dedicated method,
// not on this type, so the type stays a simple labelled enum.
type MatchStatus string

const (
	// MatchStatusDraft means the organizer is still configuring the match.
	MatchStatusDraft MatchStatus = "draft"

	// MatchStatusOpen means the match is published and accepting players.
	MatchStatusOpen MatchStatus = "open"

	// MatchStatusTeamsReady means teams have been assigned and the match
	// is ready to start.
	MatchStatusTeamsReady MatchStatus = "teams_ready"

	// MatchStatusInProgress means the match is currently being played.
	MatchStatusInProgress MatchStatus = "in_progress"

	// MatchStatusCompleted means the match has ended and post-match voting
	// is open.
	MatchStatusCompleted MatchStatus = "completed"

	// MatchStatusClosed means the match is fully closed and rankings have
	// been computed from the votes.
	MatchStatusClosed MatchStatus = "closed"
)

// ParseMatchStatus validates that the given raw value is a known match
// status and returns the typed value. Returns ErrInvalidStatus otherwise.
func ParseMatchStatus(raw string) (MatchStatus, error) {
	switch MatchStatus(raw) {
	case MatchStatusDraft,
		MatchStatusOpen,
		MatchStatusTeamsReady,
		MatchStatusInProgress,
		MatchStatusCompleted,
		MatchStatusClosed:
		return MatchStatus(raw), nil
	default:
		return "", domainerrors.ErrInvalidStatus
	}
}
