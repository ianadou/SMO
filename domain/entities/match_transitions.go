package entities

import (
	"fmt"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

// Open transitions the match from draft to open, making it visible to
// players who can then accept invitations.
func (m *Match) Open() error {
	if m.status != MatchStatusDraft {
		return fmt.Errorf("%w: cannot open match in status %q", domainerrors.ErrInvalidTransition, m.status)
	}
	m.status = MatchStatusOpen
	return nil
}

// MarkTeamsReady transitions the match from open to teams_ready, signaling
// that team assignment is complete and the match can be started.
func (m *Match) MarkTeamsReady() error {
	if m.status != MatchStatusOpen {
		return fmt.Errorf("%w: cannot mark teams ready in status %q", domainerrors.ErrInvalidTransition, m.status)
	}
	m.status = MatchStatusTeamsReady
	return nil
}

// Start transitions the match from teams_ready to in_progress, signaling
// that the match is currently being played.
func (m *Match) Start() error {
	if m.status != MatchStatusTeamsReady {
		return fmt.Errorf("%w: cannot start match in status %q", domainerrors.ErrInvalidTransition, m.status)
	}
	m.status = MatchStatusInProgress
	return nil
}

// Complete transitions the match from in_progress to completed, signaling
// that the match has ended and post-match voting is now open.
func (m *Match) Complete() error {
	if m.status != MatchStatusInProgress {
		return fmt.Errorf("%w: cannot complete match in status %q", domainerrors.ErrInvalidTransition, m.status)
	}
	m.status = MatchStatusCompleted
	return nil
}

// Close transitions the match from completed to closed, signaling that
// rankings have been computed and no further changes are allowed.
func (m *Match) Close() error {
	if m.status != MatchStatusCompleted {
		return fmt.Errorf("%w: cannot close match in status %q", domainerrors.ErrInvalidTransition, m.status)
	}
	m.status = MatchStatusClosed
	return nil
}
