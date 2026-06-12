package sharelink

import (
	"context"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// fakeShareLinkRepository stores links in a map. FindActiveByMatchID
// filters on activity at the configured reference time, like the SQL
// adapter filters on now().
type fakeShareLinkRepository struct {
	links map[entities.MatchShareLinkID]*entities.MatchShareLink
	now   time.Time

	// Optional per-method error injectors for use case error-branch
	// tests. nil = method behaves normally.
	createErr     error
	findByHashErr error
	findActiveErr error
	updateErr     error

	// findActiveOverride, when set, is returned by FindActiveByMatchID
	// regardless of its activity. It models a misbehaving adapter
	// handing back a dead link, so tests can prove the use cases defend
	// against a broken active-link invariant.
	findActiveOverride *entities.MatchShareLink
}

func newFakeShareLinkRepository(now time.Time) *fakeShareLinkRepository {
	return &fakeShareLinkRepository{
		links: make(map[entities.MatchShareLinkID]*entities.MatchShareLink),
		now:   now,
	}
}

// seedLink seeds a link on match "match-1". The match id is fixed
// because every test in the package revolves around that single match;
// only the link's own state varies.
func (r *fakeShareLinkRepository) seedLink(
	t testHelper,
	id entities.MatchShareLinkID,
	tokenHash string,
	expiresAt time.Time,
	revokedAt *time.Time,
	createdAt time.Time,
) *entities.MatchShareLink {
	link, err := entities.NewMatchShareLink(id, "match-1", tokenHash, expiresAt, revokedAt, createdAt)
	if err != nil {
		t.Fatalf("seedLink: %v", err)
	}
	r.links[id] = link
	return link
}

func (r *fakeShareLinkRepository) Create(_ context.Context, link *entities.MatchShareLink) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.links[link.ID()] = link
	return nil
}

func (r *fakeShareLinkRepository) FindByTokenHash(_ context.Context, tokenHash string) (*entities.MatchShareLink, error) {
	if r.findByHashErr != nil {
		return nil, r.findByHashErr
	}
	for _, link := range r.links {
		if link.TokenHash() == tokenHash {
			return link, nil
		}
	}
	return nil, domainerrors.ErrShareLinkNotFound
}

func (r *fakeShareLinkRepository) FindActiveByMatchID(_ context.Context, matchID entities.MatchID) (*entities.MatchShareLink, error) {
	if r.findActiveErr != nil {
		return nil, r.findActiveErr
	}
	if r.findActiveOverride != nil {
		return r.findActiveOverride, nil
	}
	for _, link := range r.links {
		if link.MatchID() == matchID && link.IsActive(r.now) {
			return link, nil
		}
	}
	return nil, domainerrors.ErrShareLinkNotFound
}

func (r *fakeShareLinkRepository) Update(_ context.Context, link *entities.MatchShareLink) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	if _, ok := r.links[link.ID()]; !ok {
		return domainerrors.ErrShareLinkNotFound
	}
	r.links[link.ID()] = link
	return nil
}
