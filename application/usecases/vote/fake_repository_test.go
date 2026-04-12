package vote

import (
	"context"
	"sync"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

type fakeVoteRepository struct {
	mu    sync.Mutex
	votes map[entities.VoteID]*entities.Vote
}

func newFakeVoteRepository() *fakeVoteRepository {
	return &fakeVoteRepository{votes: make(map[entities.VoteID]*entities.Vote)}
}

func (r *fakeVoteRepository) Save(_ context.Context, v *entities.Vote) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.votes {
		if existing.MatchID() == v.MatchID() &&
			existing.VoterID() == v.VoterID() &&
			existing.VotedID() == v.VotedID() {
			return domainerrors.ErrAlreadyVoted
		}
	}
	r.votes[v.ID()] = v
	return nil
}

func (r *fakeVoteRepository) FindByID(_ context.Context, id entities.VoteID) (*entities.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.votes[id]
	if !ok {
		return nil, domainerrors.ErrVoteNotFound
	}
	return v, nil
}

func (r *fakeVoteRepository) ListByMatch(_ context.Context, matchID entities.MatchID) ([]*entities.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*entities.Vote, 0)
	for _, v := range r.votes {
		if v.MatchID() == matchID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (r *fakeVoteRepository) ListByVoter(_ context.Context, voterID entities.PlayerID) ([]*entities.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*entities.Vote, 0)
	for _, v := range r.votes {
		if v.VoterID() == voterID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (r *fakeVoteRepository) Delete(_ context.Context, id entities.VoteID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.votes, id)
	return nil
}
