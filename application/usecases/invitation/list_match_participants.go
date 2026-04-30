package invitation

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// ListMatchParticipantsUseCase returns the confirmed participants for a
// given match. A confirmed participant is a player whose invitation has
// `used_at IS NOT NULL` (the player clicked their invite link). This is
// the primary input of the match-detail organizer view and the future
// team-generation use case.
type ListMatchParticipantsUseCase struct {
	repo ports.InvitationRepository
}

// NewListMatchParticipantsUseCase builds the use case.
func NewListMatchParticipantsUseCase(repo ports.InvitationRepository) *ListMatchParticipantsUseCase {
	return &ListMatchParticipantsUseCase{repo: repo}
}

// Execute returns the participants ordered by confirmation time ascending.
// Empty match (no confirmed invitation yet) returns an empty slice, not nil.
func (uc *ListMatchParticipantsUseCase) Execute(ctx context.Context, matchID entities.MatchID) ([]entities.MatchParticipant, error) {
	if matchID == "" {
		return nil, fmt.Errorf("list match participants use case: empty match id")
	}
	participants, err := uc.repo.ListConfirmedParticipants(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("list match participants use case: %w", err)
	}
	if participants == nil {
		participants = []entities.MatchParticipant{}
	}
	return participants, nil
}
