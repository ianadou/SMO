package vote

import (
	"context"
	"sync"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// fakeMatchRepository is a minimal MatchRepository for Vote tests. The
// read methods used by the vote use cases return configurable data;
// the mutating methods panic if accidentally called.
type fakeMatchRepository struct {
	mu      sync.Mutex
	matches map[entities.MatchID]*entities.Match

	members        []entities.MatchTeamMember
	togetherCounts map[entities.PlayerID]int
	previousMatch  *entities.Match
}

func newFakeMatchRepository() *fakeMatchRepository {
	return &fakeMatchRepository{
		matches:        make(map[entities.MatchID]*entities.Match),
		togetherCounts: make(map[entities.PlayerID]int),
	}
}

func (r *fakeMatchRepository) addMatch(m *entities.Match) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matches[m.ID()] = m
}

func (r *fakeMatchRepository) FindByID(_ context.Context, id entities.MatchID) (*entities.Match, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.matches[id]
	if !ok {
		return nil, domainerrors.ErrMatchNotFound
	}
	return m, nil
}

func (r *fakeMatchRepository) ListTeamMembersWithPlayers(_ context.Context, _ entities.MatchID) ([]entities.MatchTeamMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.members, nil
}

func (r *fakeMatchRepository) CountClosedMatchesTogether(
	_ context.Context, _ entities.GroupID, _ entities.PlayerID, _ []entities.PlayerID,
) (map[entities.PlayerID]int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.togetherCounts, nil
}

func (r *fakeMatchRepository) FindLatestDecidedByGroup(_ context.Context, _ entities.GroupID, _ entities.MatchID) (*entities.Match, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.previousMatch == nil {
		return nil, domainerrors.ErrMatchNotFound
	}
	return r.previousMatch, nil
}

func (r *fakeMatchRepository) Save(context.Context, *entities.Match) error {
	panic("not implemented in vote tests")
}

func (r *fakeMatchRepository) ListByGroup(context.Context, entities.GroupID) ([]*entities.Match, error) {
	panic("not implemented in vote tests")
}

func (r *fakeMatchRepository) UpdateStatus(context.Context, *entities.Match) error {
	panic("not implemented in vote tests")
}

func (r *fakeMatchRepository) Finalize(context.Context, *entities.Match) error {
	panic("not implemented in vote tests")
}

func (r *fakeMatchRepository) ReplaceTeams(context.Context, *entities.Match) error {
	panic("not implemented in vote tests")
}

func (r *fakeMatchRepository) Delete(context.Context, entities.MatchID) error {
	panic("not implemented in vote tests")
}
