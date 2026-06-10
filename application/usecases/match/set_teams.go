package match

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
	"github.com/ianadou/smo/domain/strategies"
)

// stubPlayerName satisfies entities.NewPlayer's non-empty-name rule for
// the throwaway players passed to ManualAssignmentStrategy, which
// validates by player ID only — the name is never read.
const stubPlayerName = "stub"

// SetTeamsUseCase replaces a match's team composition with an explicit
// organizer-provided partition of the confirmed participants.
type SetTeamsUseCase struct {
	matchRepo ports.MatchRepository
	invRepo   ports.InvitationRepository
	clock     ports.Clock
}

// NewSetTeamsUseCase builds the use case. The clock feeds the kickoff
// lock check in Match.AssignTeams, which rejects edits too close to the
// scheduled start.
func NewSetTeamsUseCase(matchRepo ports.MatchRepository, invRepo ports.InvitationRepository, clock ports.Clock) *SetTeamsUseCase {
	return &SetTeamsUseCase{matchRepo: matchRepo, invRepo: invRepo, clock: clock}
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
		stub, perr := entities.NewPlayer(p.PlayerID, match.GroupID(), stubPlayerName, entities.DefaultPlayerRanking())
		if perr != nil {
			return nil, fmt.Errorf("set teams use case: build player stub: %w", perr)
		}
		players = append(players, stub)
	}

	// ManualAssignmentStrategy enforces the exact-partition rule (teams
	// must be exactly the confirmed set); Match.AssignTeams then enforces
	// status, balance and overlap. The overlap check is intentionally
	// duplicated across both — defense-in-depth at the domain boundary.
	manual, err := strategies.NewManualAssignmentStrategy(teamA, teamB)
	if err != nil {
		return nil, fmt.Errorf("set teams use case: build manual: %w", err)
	}
	resolvedA, resolvedB, err := manual.Assign(players, nil)
	if err != nil {
		return nil, fmt.Errorf("set teams use case: validate partition: %w", err)
	}

	if err := match.AssignTeams(resolvedA, resolvedB, uc.clock.Now()); err != nil {
		return nil, fmt.Errorf("set teams use case: store teams: %w", err)
	}
	if err := uc.matchRepo.ReplaceTeams(ctx, match); err != nil {
		return nil, fmt.Errorf("set teams use case: persist teams: %w", err)
	}
	return match, nil
}
