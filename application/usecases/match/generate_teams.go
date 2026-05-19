package match

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
	"github.com/ianadou/smo/domain/strategies"
)

// GenerateTeamsUseCase assigns the confirmed participants of an open
// match into two teams using the named strategy and persists the result.
type GenerateTeamsUseCase struct {
	matchRepo  ports.MatchRepository
	invRepo    ports.InvitationRepository
	playerRepo ports.PlayerRepository
	clock      ports.Clock
}

// NewGenerateTeamsUseCase builds the use case.
func NewGenerateTeamsUseCase(
	matchRepo ports.MatchRepository,
	invRepo ports.InvitationRepository,
	playerRepo ports.PlayerRepository,
	clock ports.Clock,
) *GenerateTeamsUseCase {
	return &GenerateTeamsUseCase{
		matchRepo:  matchRepo,
		invRepo:    invRepo,
		playerRepo: playerRepo,
		clock:      clock,
	}
}

// Execute resolves the confirmed participants, runs the strategy, stores
// the resulting composition on the match, and persists it.
func (uc *GenerateTeamsUseCase) Execute(
	ctx context.Context, matchID entities.MatchID, strategyName string,
) (*entities.Match, error) {
	match, err := uc.matchRepo.FindByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("generate teams use case: find match: %w", err)
	}

	participants, err := uc.invRepo.ListConfirmedParticipants(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("generate teams use case: list confirmed: %w", err)
	}

	players := make([]*entities.Player, 0, len(participants))
	for _, p := range participants {
		player, perr := uc.playerRepo.FindByID(ctx, p.PlayerID)
		if perr != nil {
			return nil, fmt.Errorf("generate teams use case: load player %q: %w", p.PlayerID, perr)
		}
		players = append(players, player)
	}

	strategy, err := uc.strategyByName(strategyName)
	if err != nil {
		return nil, err
	}

	previousWinner, err := uc.previousWinner(ctx, match)
	if err != nil {
		return nil, err
	}

	teamA, teamB, err := strategy.Assign(players, previousWinner)
	if err != nil {
		return nil, fmt.Errorf("generate teams use case: assign: %w", err)
	}

	if err := match.AssignTeams(teamA, teamB); err != nil {
		return nil, fmt.Errorf("generate teams use case: store teams: %w", err)
	}

	if err := uc.matchRepo.ReplaceTeams(ctx, match); err != nil {
		return nil, fmt.Errorf("generate teams use case: persist teams: %w", err)
	}
	return match, nil
}

// previousWinner resolves the side that won the group's most recent
// decided match. No prior decided match (ErrMatchNotFound) is the
// expected first-match case and yields nil, not an error.
func (uc *GenerateTeamsUseCase) previousWinner(ctx context.Context, m *entities.Match) (*entities.TeamSide, error) {
	previous, err := uc.matchRepo.FindLatestDecidedByGroup(ctx, m.GroupID(), m.ID())
	if errors.Is(err, domainerrors.ErrMatchNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("generate teams use case: find previous match: %w", err)
	}
	return previous.WinningSide(), nil
}

func (uc *GenerateTeamsUseCase) strategyByName(name string) (strategies.AssignmentStrategy, error) {
	switch name {
	case "random":
		seed := uint64(uc.clock.Now().UnixNano())
		//nolint:gosec // team shuffling is a game mechanic, not a security primitive; PRNG is intentional and seeded for reproducibility
		return strategies.NewRandomAssignmentStrategy(rand.New(rand.NewPCG(seed, seed)))
	case "ranking":
		return strategies.NewRankingBasedStrategy(), nil
	default:
		return nil, fmt.Errorf("%w: unknown strategy %q", domainerrors.ErrInvalidParameter, name)
	}
}
