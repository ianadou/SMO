package match

import (
	"context"
	"sync"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

type fakeMatchRepository struct {
	mu          sync.Mutex
	matches     map[entities.MatchID]*entities.Match
	teamMembers map[entities.MatchID][]entities.MatchTeamMember

	// Optional per-method error injectors for use case error-branch
	// tests. nil = method behaves normally; non-nil = method returns
	// the configured error verbatim.
	saveErr         error
	findByIDErr     error
	listByGroupErr  error
	updateStatusErr error
	finalizeErr     error
	replaceTeamsErr error

	// Call counters for assertions that a port method was actually
	// invoked (guards against vacuous shared-pointer-mutation tests).
	replaceTeamsCalls int

	// latestDecided is returned by FindLatestDecidedByGroup. nil means
	// "no previous decided match" (the use case treats ErrMatchNotFound
	// as the first-match case), so the default keeps the snake draft
	// unbiased unless a test wires a previous result.
	latestDecided *entities.Match
}

func newFakeMatchRepository() *fakeMatchRepository {
	return &fakeMatchRepository{matches: make(map[entities.MatchID]*entities.Match)}
}

func (r *fakeMatchRepository) Save(_ context.Context, m *entities.Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.saveErr != nil {
		return r.saveErr
	}
	r.matches[m.ID()] = m
	return nil
}

func (r *fakeMatchRepository) FindByID(_ context.Context, id entities.MatchID) (*entities.Match, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.findByIDErr != nil {
		return nil, r.findByIDErr
	}
	m, ok := r.matches[id]
	if !ok {
		return nil, domainerrors.ErrMatchNotFound
	}
	return m, nil
}

func (r *fakeMatchRepository) ListByGroup(_ context.Context, groupID entities.GroupID) ([]*entities.Match, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.listByGroupErr != nil {
		return nil, r.listByGroupErr
	}
	result := make([]*entities.Match, 0)
	for _, m := range r.matches {
		if m.GroupID() == groupID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *fakeMatchRepository) UpdateStatus(_ context.Context, m *entities.Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.updateStatusErr != nil {
		return r.updateStatusErr
	}
	if _, exists := r.matches[m.ID()]; !exists {
		return domainerrors.ErrMatchNotFound
	}
	r.matches[m.ID()] = m
	return nil
}

func (r *fakeMatchRepository) Finalize(_ context.Context, m *entities.Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.finalizeErr != nil {
		return r.finalizeErr
	}
	if _, exists := r.matches[m.ID()]; !exists {
		return domainerrors.ErrMatchNotFound
	}
	r.matches[m.ID()] = m
	return nil
}

func (r *fakeMatchRepository) ReplaceTeams(_ context.Context, m *entities.Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.replaceTeamsCalls++
	if r.replaceTeamsErr != nil {
		return r.replaceTeamsErr
	}
	if _, exists := r.matches[m.ID()]; !exists {
		return domainerrors.ErrMatchNotFound
	}
	r.matches[m.ID()] = m
	return nil
}

func (r *fakeMatchRepository) ListTeamMembersWithPlayers(_ context.Context, matchID entities.MatchID) ([]entities.MatchTeamMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.teamMembers[matchID], nil
}

func (r *fakeMatchRepository) CountClosedMatchesTogether(context.Context, entities.GroupID, entities.PlayerID, []entities.PlayerID) (map[entities.PlayerID]int, error) {
	return map[entities.PlayerID]int{}, nil
}

func (r *fakeMatchRepository) FindLatestDecidedByGroup(_ context.Context, _ entities.GroupID, _ entities.MatchID) (*entities.Match, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.latestDecided == nil {
		return nil, domainerrors.ErrMatchNotFound
	}
	return r.latestDecided, nil
}

func (r *fakeMatchRepository) Delete(_ context.Context, id entities.MatchID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.matches, id)
	return nil
}
