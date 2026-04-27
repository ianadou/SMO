package match

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
	"github.com/ianadou/smo/domain/ranking"
)

// FinalizeMatchUseCase orchestrates the closing of a completed match:
// it elects the MVP from the votes, recomputes the rankings of every
// player who appears in those votes, persists the new rankings, and
// transitions the match from completed to closed in a single repository
// call (see MatchRepository.Finalize).
//
// The use case is idempotent: the ranking calculation is a pure function
// of the votes, and the votes are immutable once a match is completed,
// so re-running Execute on a (rare) retry converges to the same state.
type FinalizeMatchUseCase struct {
	matchRepo  ports.MatchRepository
	voteRepo   ports.VoteRepository
	playerRepo ports.PlayerRepository
	calculator *ranking.Calculator
}

// FinalizeMatchOutput is the result returned by Execute.
type FinalizeMatchOutput struct {
	Match           *entities.Match
	MVP             *entities.PlayerID
	UpdatedRankings map[entities.PlayerID]int
}

// NewFinalizeMatchUseCase builds a FinalizeMatchUseCase.
func NewFinalizeMatchUseCase(
	matchRepo ports.MatchRepository,
	voteRepo ports.VoteRepository,
	playerRepo ports.PlayerRepository,
	calculator *ranking.Calculator,
) *FinalizeMatchUseCase {
	return &FinalizeMatchUseCase{
		matchRepo:  matchRepo,
		voteRepo:   voteRepo,
		playerRepo: playerRepo,
		calculator: calculator,
	}
}

// Execute finalizes the match with the given ID.
func (uc *FinalizeMatchUseCase) Execute(ctx context.Context, id entities.MatchID) (*FinalizeMatchOutput, error) {
	match, err := uc.matchRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finalize match use case: find match %q: %w", id, err)
	}

	votes, err := uc.voteRepo.ListByMatch(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finalize match use case: list votes for match %q: %w", id, err)
	}

	participants, err := uc.loadParticipants(ctx, votes)
	if err != nil {
		return nil, fmt.Errorf("finalize match use case: load participants for match %q: %w", id, err)
	}

	mvp := selectMVP(votes)

	newRankings, err := uc.calculator.Compute(participants, votes)
	if err != nil {
		return nil, fmt.Errorf("finalize match use case: compute rankings for match %q: %w", id, err)
	}

	if applyErr := uc.applyRankings(ctx, participants, newRankings); applyErr != nil {
		return nil, fmt.Errorf("finalize match use case: apply rankings for match %q: %w", id, applyErr)
	}

	if transErr := match.Finalize(mvp); transErr != nil {
		return nil, fmt.Errorf("finalize match use case: transition match %q: %w", id, transErr)
	}

	if saveErr := uc.matchRepo.Finalize(ctx, match); saveErr != nil {
		return nil, fmt.Errorf("finalize match use case: persist match %q: %w", id, saveErr)
	}

	return &FinalizeMatchOutput{
		Match:           match,
		MVP:             mvp,
		UpdatedRankings: newRankings,
	}, nil
}

// loadParticipants resolves the unique players appearing in the votes
// (as voter or voted) into hydrated Player entities. Players who
// participated but neither voted nor were voted for are not loaded:
// the ranking calculator would not modify their score anyway.
func (uc *FinalizeMatchUseCase) loadParticipants(
	ctx context.Context,
	votes []*entities.Vote,
) ([]*entities.Player, error) {
	uniqueIDs := make(map[entities.PlayerID]struct{}, len(votes)*2)
	for _, vote := range votes {
		uniqueIDs[vote.VoterID()] = struct{}{}
		uniqueIDs[vote.VotedID()] = struct{}{}
	}

	players := make([]*entities.Player, 0, len(uniqueIDs))
	for playerID := range uniqueIDs {
		player, err := uc.playerRepo.FindByID(ctx, playerID)
		if err != nil {
			return nil, fmt.Errorf("find player %q: %w", playerID, err)
		}
		players = append(players, player)
	}
	return players, nil
}

// applyRankings rebuilds each participant with their new ranking and
// persists the change. The Player entity is immutable except through
// its constructor, so we rebuild rather than mutate.
func (uc *FinalizeMatchUseCase) applyRankings(
	ctx context.Context,
	participants []*entities.Player,
	newRankings map[entities.PlayerID]int,
) error {
	for _, player := range participants {
		newRanking, ok := newRankings[player.ID()]
		if !ok || newRanking == player.Ranking() {
			continue
		}
		updated, err := entities.NewPlayer(player.ID(), player.GroupID(), player.Name(), newRanking)
		if err != nil {
			return fmt.Errorf("rebuild player %q with new ranking: %w", player.ID(), err)
		}
		if updateErr := uc.playerRepo.UpdateRanking(ctx, updated); updateErr != nil {
			return fmt.Errorf("persist new ranking for player %q: %w", player.ID(), updateErr)
		}
	}
	return nil
}
