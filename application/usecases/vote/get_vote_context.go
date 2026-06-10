package vote

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
)

// PageContext is the assembled, display-oriented view a player needs to
// render the vote page in every lifecycle state: rate teammates while
// the match is completed, watch collective progress once their own
// votes are in, and read the final aggregates once the match is closed.
//
// Names are kept at the application boundary, mirroring the invitation
// PageContext: the HTTP layer decides what the unauthenticated token
// bearer may see (initials are derived there, raw votes never leave the
// backend — only aggregates do).
type PageContext struct {
	GroupName   string
	MatchTitle  string
	Venue       string
	ScheduledAt time.Time
	Status      entities.MatchStatus
	ScoreA      *int
	ScoreB      *int
	WinningSide *entities.TeamSide

	// Voter and Teammates are zero-valued until the match reaches
	// Completed: before that there is nothing to rate and teams may not
	// even be assigned yet.
	Voter       Voter
	Teammates   []RateableTeammate
	VotersDone  int
	VotersTotal int

	// Results is non-nil only when the match is Closed.
	Results *Results
}

// Voter identifies the token bearer inside the vote context.
type Voter struct {
	PlayerID entities.PlayerID
	Name     string
	Team     entities.TeamSide
}

// RateableTeammate is one row of the rating list (view A of the kit).
// YourScore carries the bearer's already-cast score so the page can
// lock that row, or nil when the teammate has not been rated yet.
type RateableTeammate struct {
	PlayerID        entities.PlayerID
	Name            string
	MatchesTogether int
	YourScore       *int
}

// Results carries the closed-match aggregates (view D of the kit).
type Results struct {
	Teammates []TeammateResult
	Self      SelfResult
}

// TeammateResult is one teammate's final aggregate. Delta is the
// difference with the player's average on the group's previous decided
// match, nil when that match does not exist or the player received no
// vote in it.
type TeammateResult struct {
	PlayerID   entities.PlayerID
	Name       string
	Average    float64
	VotesCount int
	Delta      *float64
}

// SelfResult is the bearer's own aggregate. Average is nil when nobody
// rated the bearer.
type SelfResult struct {
	Average    *float64
	VotesCount int
}

// GetVoteContextUseCase resolves a plain invitation token into the full
// vote-page context. It is a pure read and validates only two things:
// the token resolves to an invitation, and the bearer confirmed
// attendance. Every other state (match not completed yet, already
// voted, closed) is reported in the context, not rejected — the page
// decides which screen to show.
type GetVoteContextUseCase struct {
	invRepo   ports.InvitationRepository
	matchRepo ports.MatchRepository
	groupRepo ports.GroupRepository
	voteRepo  ports.VoteRepository
	tokens    ports.InvitationTokenService
}

// NewGetVoteContextUseCase builds the use case.
func NewGetVoteContextUseCase(
	invRepo ports.InvitationRepository,
	matchRepo ports.MatchRepository,
	groupRepo ports.GroupRepository,
	voteRepo ports.VoteRepository,
	tokens ports.InvitationTokenService,
) *GetVoteContextUseCase {
	return &GetVoteContextUseCase{
		invRepo:   invRepo,
		matchRepo: matchRepo,
		groupRepo: groupRepo,
		voteRepo:  voteRepo,
		tokens:    tokens,
	}
}

// Execute assembles the context for the bearer of the given plain
// token. Returns ErrInvitationNotFound for an unknown token and
// ErrNotConfirmedParticipant when the bearer declined or never answered.
func (uc *GetVoteContextUseCase) Execute(ctx context.Context, plainToken string) (*PageContext, error) {
	invitation, err := uc.invRepo.FindByTokenHash(ctx, uc.tokens.HashToken(plainToken))
	if err != nil {
		return nil, fmt.Errorf("get vote context use case: find invitation by hash: %w", err)
	}
	if !invitation.IsConfirmed() {
		return nil, fmt.Errorf("get vote context use case: bearer declined or never answered: %w",
			domainerrors.ErrNotConfirmedParticipant)
	}

	match, err := uc.matchRepo.FindByID(ctx, invitation.MatchID())
	if err != nil {
		return nil, fmt.Errorf("get vote context use case: find match: %w", err)
	}

	group, err := uc.groupRepo.FindByID(ctx, match.GroupID())
	if err != nil {
		return nil, fmt.Errorf("get vote context use case: find group: %w", err)
	}

	pageContext := &PageContext{
		GroupName:   group.Name(),
		MatchTitle:  match.Title(),
		Venue:       match.Venue(),
		ScheduledAt: match.ScheduledAt(),
		Status:      match.Status(),
		ScoreA:      match.ScoreA(),
		ScoreB:      match.ScoreB(),
		WinningSide: match.WinningSide(),
	}

	if match.Status() != entities.MatchStatusCompleted && match.Status() != entities.MatchStatusClosed {
		// Nothing to rate yet: the page only needs the match block to
		// render its "voting opens after the match" state.
		return pageContext, nil
	}

	if err := uc.fillRatingView(ctx, pageContext, match, invitation.PlayerID()); err != nil {
		return nil, err
	}

	if match.Status() == entities.MatchStatusClosed {
		if err := uc.fillResults(ctx, pageContext, match, invitation.PlayerID()); err != nil {
			return nil, err
		}
	}

	return pageContext, nil
}

// fillRatingView loads the roster and votes to populate the voter
// block, the rateable teammates (with the bearer's own scores) and the
// collective progress counters.
func (uc *GetVoteContextUseCase) fillRatingView(
	ctx context.Context,
	pageContext *PageContext,
	match *entities.Match,
	voterID entities.PlayerID,
) error {
	members, err := uc.matchRepo.ListTeamMembersWithPlayers(ctx, match.ID())
	if err != nil {
		return fmt.Errorf("get vote context use case: list team members: %w", err)
	}
	nameByPlayer := make(map[entities.PlayerID]string, len(members))
	for _, member := range members {
		nameByPlayer[member.PlayerID] = member.PlayerName
	}

	voterSide, voterAssigned := match.TeamOf(voterID)
	if !voterAssigned {
		// A confirmed participant on a completed match is always on a
		// roster (exact-partition rule); a miss is a data-integrity bug.
		return fmt.Errorf("get vote context use case: voter %q missing from rosters: %w",
			voterID, domainerrors.ErrPlayerNotInMatch)
	}
	pageContext.Voter = Voter{PlayerID: voterID, Name: nameByPlayer[voterID], Team: voterSide}

	votes, err := uc.voteRepo.ListByMatch(ctx, match.ID())
	if err != nil {
		return fmt.Errorf("get vote context use case: list votes: %w", err)
	}

	yourScoreByVoted := make(map[entities.PlayerID]int)
	distinctVoters := make(map[entities.PlayerID]struct{})
	for _, v := range votes {
		distinctVoters[v.VoterID()] = struct{}{}
		if v.VoterID() == voterID {
			score := v.Score()
			yourScoreByVoted[v.VotedID()] = score
		}
	}

	teammateIDs := match.TeammatesOf(voterID)
	matchesTogether, err := uc.matchRepo.CountClosedMatchesTogether(ctx, match.GroupID(), voterID, teammateIDs)
	if err != nil {
		return fmt.Errorf("get vote context use case: count matches together: %w", err)
	}

	teammates := make([]RateableTeammate, 0, len(teammateIDs))
	for _, teammateID := range teammateIDs {
		teammate := RateableTeammate{
			PlayerID:        teammateID,
			Name:            nameByPlayer[teammateID],
			MatchesTogether: matchesTogether[teammateID],
		}
		if score, voted := yourScoreByVoted[teammateID]; voted {
			teammate.YourScore = &score
		}
		teammates = append(teammates, teammate)
	}

	pageContext.Teammates = teammates
	pageContext.VotersDone = len(distinctVoters)
	pageContext.VotersTotal = len(members)
	return nil
}

// fillResults computes the closed-match aggregates for the bearer's
// team: per-teammate average and vote count, the delta against the
// group's previous decided match, and the bearer's own line.
func (uc *GetVoteContextUseCase) fillResults(
	ctx context.Context,
	pageContext *PageContext,
	match *entities.Match,
	voterID entities.PlayerID,
) error {
	votes, err := uc.voteRepo.ListByMatch(ctx, match.ID())
	if err != nil {
		return fmt.Errorf("get vote context use case: list votes for results: %w", err)
	}
	averages, counts := aggregateByVoted(votes)

	previousAverages, err := uc.previousMatchAverages(ctx, match)
	if err != nil {
		return err
	}

	results := &Results{Teammates: make([]TeammateResult, 0, len(pageContext.Teammates))}
	for _, teammate := range pageContext.Teammates {
		result := TeammateResult{
			PlayerID:   teammate.PlayerID,
			Name:       teammate.Name,
			Average:    averages[teammate.PlayerID],
			VotesCount: counts[teammate.PlayerID],
		}
		if previous, rated := previousAverages[teammate.PlayerID]; rated {
			delta := result.Average - previous
			result.Delta = &delta
		}
		results.Teammates = append(results.Teammates, result)
	}

	results.Self = SelfResult{VotesCount: counts[voterID]}
	if results.Self.VotesCount > 0 {
		average := averages[voterID]
		results.Self.Average = &average
	}

	pageContext.Results = results
	return nil
}

// previousMatchAverages returns the per-player vote averages of the
// group's previous decided match, or an empty map when no such match
// exists (the normal first-match case).
func (uc *GetVoteContextUseCase) previousMatchAverages(
	ctx context.Context,
	match *entities.Match,
) (map[entities.PlayerID]float64, error) {
	previous, err := uc.matchRepo.FindLatestDecidedByGroup(ctx, match.GroupID(), match.ID())
	if errors.Is(err, domainerrors.ErrMatchNotFound) {
		return map[entities.PlayerID]float64{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get vote context use case: find previous match: %w", err)
	}

	previousVotes, err := uc.voteRepo.ListByMatch(ctx, previous.ID())
	if err != nil {
		return nil, fmt.Errorf("get vote context use case: list previous votes: %w", err)
	}
	averages, _ := aggregateByVoted(previousVotes)
	return averages, nil
}

// aggregateByVoted reduces raw votes into per-player average and count.
func aggregateByVoted(votes []*entities.Vote) (map[entities.PlayerID]float64, map[entities.PlayerID]int) {
	totals := make(map[entities.PlayerID]int)
	counts := make(map[entities.PlayerID]int)
	for _, v := range votes {
		totals[v.VotedID()] += v.Score()
		counts[v.VotedID()]++
	}

	averages := make(map[entities.PlayerID]float64, len(totals))
	for playerID, total := range totals {
		averages[playerID] = float64(total) / float64(counts[playerID])
	}
	return averages, counts
}
