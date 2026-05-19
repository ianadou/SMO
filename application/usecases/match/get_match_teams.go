package match

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// GetMatchTeamsUseCase returns the display-ready team membership of a
// match (player id + name + team + slot), ordered by team then slot.
type GetMatchTeamsUseCase struct {
	matchRepo ports.MatchRepository
}

// NewGetMatchTeamsUseCase builds the use case.
func NewGetMatchTeamsUseCase(matchRepo ports.MatchRepository) *GetMatchTeamsUseCase {
	return &GetMatchTeamsUseCase{matchRepo: matchRepo}
}

// Execute loads the joined membership rows.
func (uc *GetMatchTeamsUseCase) Execute(ctx context.Context, matchID entities.MatchID) ([]entities.MatchTeamMember, error) {
	members, err := uc.matchRepo.ListTeamMembersWithPlayers(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("get match teams use case: %w", err)
	}
	return members, nil
}
