package sharelink

import (
	"context"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// fakeMatchRepository is a minimal MatchRepository tailored to the
// share link use cases: only FindByID is exercised. Tests register the
// matches they expect to be looked up.
type fakeMatchRepository struct {
	matches map[entities.MatchID]*entities.Match
}

func newFakeMatchRepository() *fakeMatchRepository {
	return &fakeMatchRepository{matches: make(map[entities.MatchID]*entities.Match)}
}

func (r *fakeMatchRepository) seedMatch(t testHelper, id entities.MatchID, groupID entities.GroupID) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	scheduled := now.Add(2 * 24 * time.Hour)
	m, err := entities.NewMatch(id, groupID, "Match", "Venue", scheduled, now)
	if err != nil {
		t.Fatalf("seedMatch: %v", err)
	}
	r.matches[id] = m
}

// seedMatchWithStatus seeds match "match-1" (group "group-1")
// rehydrated directly into the given status so share link tests can
// drive the attendance-lock branch. The ids are fixed because every
// caller uses the same pair; only the status varies.
func (r *fakeMatchRepository) seedMatchWithStatus(t testHelper, status entities.MatchStatus) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	scheduled := now.Add(2 * 24 * time.Hour)
	m, err := entities.RehydrateMatch(entities.MatchSnapshot{
		ID:          "match-1",
		GroupID:     "group-1",
		Title:       "Match",
		Venue:       "Venue",
		ScheduledAt: scheduled,
		Status:      status,
		CreatedAt:   now,
	})
	if err != nil {
		t.Fatalf("seedMatchWithStatus: %v", err)
	}
	r.matches["match-1"] = m
}

func (r *fakeMatchRepository) FindByID(_ context.Context, id entities.MatchID) (*entities.Match, error) {
	m, ok := r.matches[id]
	if !ok {
		return nil, domainerrors.ErrMatchNotFound
	}
	return m, nil
}

func (r *fakeMatchRepository) Save(context.Context, *entities.Match) error         { return nil }
func (r *fakeMatchRepository) UpdateStatus(context.Context, *entities.Match) error { return nil }
func (r *fakeMatchRepository) Finalize(context.Context, *entities.Match) error     { return nil }
func (r *fakeMatchRepository) ReplaceTeams(context.Context, *entities.Match) error { return nil }
func (r *fakeMatchRepository) Delete(context.Context, entities.MatchID) error      { return nil }

func (r *fakeMatchRepository) ListByGroup(context.Context, entities.GroupID) ([]*entities.Match, error) {
	return nil, nil
}

func (r *fakeMatchRepository) ListTeamMembersWithPlayers(context.Context, entities.MatchID) ([]entities.MatchTeamMember, error) {
	return nil, nil
}

func (r *fakeMatchRepository) CountClosedMatchesTogether(context.Context, entities.GroupID, entities.PlayerID, []entities.PlayerID) (map[entities.PlayerID]int, error) {
	return map[entities.PlayerID]int{}, nil
}

func (r *fakeMatchRepository) FindLatestDecidedByGroup(context.Context, entities.GroupID, entities.MatchID) (*entities.Match, error) {
	return nil, domainerrors.ErrMatchNotFound
}

// fakePlayerRepository is a minimal PlayerRepository for share link use
// case tests: FindByID, ListByGroup and Save are exercised; the rest of
// the port is stubbed.
type fakePlayerRepository struct {
	players map[entities.PlayerID]*entities.Player

	// Per-method error injectors for use case error-branch tests.
	// nil = method behaves normally.
	saveErr error
	findErr error
	listErr error
}

func newFakePlayerRepository() *fakePlayerRepository {
	return &fakePlayerRepository{players: make(map[entities.PlayerID]*entities.Player)}
}

// seedPlayer seeds a player in group "group-1". The group id is fixed
// because every test in the package revolves around that single group.
func (r *fakePlayerRepository) seedPlayer(t testHelper, id entities.PlayerID, name string) {
	p, err := entities.NewPlayer(id, "group-1", name, 1000)
	if err != nil {
		t.Fatalf("seedPlayer: %v", err)
	}
	r.players[id] = p
}

func (r *fakePlayerRepository) Save(_ context.Context, player *entities.Player) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.players[player.ID()] = player
	return nil
}

func (r *fakePlayerRepository) FindByID(_ context.Context, id entities.PlayerID) (*entities.Player, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	p, ok := r.players[id]
	if !ok {
		return nil, domainerrors.ErrPlayerNotFound
	}
	return p, nil
}

func (r *fakePlayerRepository) ListByGroup(_ context.Context, groupID entities.GroupID) ([]*entities.Player, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	result := make([]*entities.Player, 0)
	for _, p := range r.players {
		if p.GroupID() == groupID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *fakePlayerRepository) UpdateRanking(context.Context, *entities.Player) error { return nil }
func (r *fakePlayerRepository) Delete(context.Context, entities.PlayerID) error       { return nil }

// testHelper is the minimal interface satisfied by *testing.T.
type testHelper interface {
	Fatalf(format string, args ...any)
}
