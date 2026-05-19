package match

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
	"github.com/ianadou/smo/domain/strategies"
)

// SetTeamsUseCase replaces a match's team composition with an explicit
// organizer-provided partition of the confirmed participants.
type SetTeamsUseCase struct {
	matchRepo ports.MatchRepository
	invRepo   ports.InvitationRepository
}

// NewSetTeamsUseCase builds the use case.
func NewSetTeamsUseCase(matchRepo ports.MatchRepository, invRepo ports.InvitationRepository) *SetTeamsUseCase {
	return &SetTeamsUseCase{matchRepo: matchRepo, invRepo: invRepo}
}

// Execute validates that (teamA ∪ teamB) is exactly the set of confirmed
// participants, then stores and persists the composition.
func (uc *SetTeamsUseCase) Execute(ctx context.Context, matchID entities.MatchID, teamA, teamB []entities.PlayerID) (*entities.Match, error) {
	match, err := uc.matchRepo.FindByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("set teams use case: find match: %w", err)
	}

	participants, err := uc.invRepo.ListConfirmedParticipants(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("set teams use case: list confirmed: %w", err)
	}
	players := make([]*entities.Player, 0, len(participants))
	for _, p := range participants {
		stub, perr := entities.NewPlayer(p.PlayerID, match.GroupID(), "x", entities.DefaultPlayerRanking())
		if perr != nil {
			return nil, fmt.Errorf("set teams use case: build player stub: %w", perr)
		}
		players = append(players, stub)
	}

	manual, err := strategies.NewManualAssignmentStrategy(teamA, teamB)
	if err != nil {
		return nil, fmt.Errorf("set teams use case: build manual: %w", err)
	}
	resolvedA, resolvedB, err := manual.Assign(players)
	if err != nil {
		return nil, fmt.Errorf("set teams use case: validate partition: %w", err)
	}

	if err := match.AssignTeams(resolvedA, resolvedB); err != nil {
		return nil, fmt.Errorf("set teams use case: store teams: %w", err)
	}
	if err := uc.matchRepo.ReplaceTeams(ctx, match); err != nil {
		return nil, fmt.Errorf("set teams use case: persist teams: %w", err)
	}
	return match, nil
}
