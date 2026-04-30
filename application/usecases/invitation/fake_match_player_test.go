package invitation

import (
	"context"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// fakeMatchRepo is a minimal MatchRepository tailored to the invitation
// use cases: it only supports FindByID. Tests register the matches they
// expect to be looked up.
type fakeMatchRepo struct {
	matches map[entities.MatchID]*entities.Match
}

func newFakeMatchRepo() *fakeMatchRepo {
	return &fakeMatchRepo{matches: make(map[entities.MatchID]*entities.Match)}
}

func (r *fakeMatchRepo) seedMatch(t testHelper, id entities.MatchID, groupID entities.GroupID) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	scheduled := now.Add(2 * 24 * time.Hour)
	m, err := entities.NewMatch(id, groupID, "Match", "Venue", scheduled, now)
	if err != nil {
		t.Fatalf("seedMatch: %v", err)
	}
	r.matches[id] = m
}

func (r *fakeMatchRepo) FindByID(_ context.Context, id entities.MatchID) (*entities.Match, error) {
	m, ok := r.matches[id]
	if !ok {
		return nil, domainerrors.ErrMatchNotFound
	}
	return m, nil
}

func (r *fakeMatchRepo) Save(context.Context, *entities.Match) error         { return nil }
func (r *fakeMatchRepo) UpdateStatus(context.Context, *entities.Match) error { return nil }
func (r *fakeMatchRepo) Finalize(context.Context, *entities.Match) error     { return nil }
func (r *fakeMatchRepo) Delete(context.Context, entities.MatchID) error      { return nil }
func (r *fakeMatchRepo) ListByGroup(context.Context, entities.GroupID) ([]*entities.Match, error) {
	return nil, nil
}

// fakePlayerRepo is a minimal PlayerRepository for invitation use case
// tests. Only FindByID is exercised; the rest of the port is stubbed.
type fakePlayerRepo struct {
	players map[entities.PlayerID]*entities.Player
}

func newFakePlayerRepo() *fakePlayerRepo {
	return &fakePlayerRepo{players: make(map[entities.PlayerID]*entities.Player)}
}

func (r *fakePlayerRepo) seedPlayer(t testHelper, id entities.PlayerID, groupID entities.GroupID) {
	p, err := entities.NewPlayer(id, groupID, "Test Player", 1000)
	if err != nil {
		t.Fatalf("seedPlayer: %v", err)
	}
	r.players[id] = p
}

func (r *fakePlayerRepo) FindByID(_ context.Context, id entities.PlayerID) (*entities.Player, error) {
	p, ok := r.players[id]
	if !ok {
		return nil, domainerrors.ErrPlayerNotFound
	}
	return p, nil
}

func (r *fakePlayerRepo) Save(context.Context, *entities.Player) error          { return nil }
func (r *fakePlayerRepo) UpdateRanking(context.Context, *entities.Player) error { return nil }
func (r *fakePlayerRepo) Delete(context.Context, entities.PlayerID) error       { return nil }
func (r *fakePlayerRepo) ListByGroup(context.Context, entities.GroupID) ([]*entities.Player, error) {
	return nil, nil
}

// testHelper is the minimal interface satisfied by *testing.T.
type testHelper interface {
	Fatalf(format string, args ...any)
}
