//go:build integration

package repositories_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/repositories"
)

// newTestVoteRepository cleans the votes table and re-seeds a match +
// 2 players so vote rows can reference valid FKs.
func newTestVoteRepository(t *testing.T) *repositories.PostgresVoteRepository {
	t.Helper()
	ctx := context.Background()

	if _, err := sharedPool.Exec(ctx, "DELETE FROM votes"); err != nil {
		t.Fatalf("clean votes: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, "DELETE FROM matches"); err != nil {
		t.Fatalf("clean matches: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, "DELETE FROM players"); err != nil {
		t.Fatalf("clean players: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, `
		INSERT INTO groups (id, organizer_id, name)
		VALUES ('test-group', 'test-org', 'Test Group')
		ON CONFLICT (id) DO NOTHING
	`); err != nil {
		t.Fatalf("seed group: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, `
		INSERT INTO matches (id, group_id, title, venue, scheduled_at, status)
		VALUES ('test-match', 'test-group', 'Test', 'V', NOW() + INTERVAL '1 day', 'completed')
	`); err != nil {
		t.Fatalf("seed match: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, `
		INSERT INTO players (id, group_id, name, ranking)
		VALUES ('p-voter', 'test-group', 'Voter', 1000),
		       ('p-voted', 'test-group', 'Voted', 1000)
	`); err != nil {
		t.Fatalf("seed players: %v", err)
	}
	return repositories.NewPostgresVoteRepository(sharedPool)
}

func buildTestVote(t *testing.T, id string, voter, voted entities.PlayerID, score int) *entities.Vote {
	t.Helper()
	v, err := entities.NewVote(entities.VoteID(id), "test-match", voter, voted, score, time.Now())
	if err != nil {
		t.Fatalf("NewVote: %v", err)
	}
	return v
}

func TestPostgresVoteRepository_Save_Persists(t *testing.T) {
	repo := newTestVoteRepository(t)
	v := buildTestVote(t, "v-1", "p-voter", "p-voted", 4)
	if err := repo.Save(context.Background(), v); err != nil {
		t.Fatalf("save: %v", err)
	}
	found, _ := repo.FindByID(context.Background(), "v-1")
	if found.Score() != 4 {
		t.Errorf("expected 4, got %d", found.Score())
	}
}

func TestPostgresVoteRepository_Save_ReturnsErrAlreadyVoted_OnUniqueViolation(t *testing.T) {
	repo := newTestVoteRepository(t)
	ctx := context.Background()
	v1 := buildTestVote(t, "v-1", "p-voter", "p-voted", 4)
	v2 := buildTestVote(t, "v-2", "p-voter", "p-voted", 5) // same triplet

	if err := repo.Save(ctx, v1); err != nil {
		t.Fatalf("first save: %v", err)
	}
	err := repo.Save(ctx, v2)
	if !errors.Is(err, domainerrors.ErrAlreadyVoted) {
		t.Errorf("expected ErrAlreadyVoted, got %v", err)
	}
}

func TestPostgresVoteRepository_Save_ReturnsErrReferencedEntityNotFound(t *testing.T) {
	repo := newTestVoteRepository(t)
	v := buildTestVote(t, "v-1", "nonexistent-voter", "p-voted", 4)
	err := repo.Save(context.Background(), v)
	if !errors.Is(err, domainerrors.ErrReferencedEntityNotFound) {
		t.Errorf("expected ErrReferencedEntityNotFound, got %v", err)
	}
}

func TestPostgresVoteRepository_FindByID_ReturnsErrVoteNotFound(t *testing.T) {
	repo := newTestVoteRepository(t)
	_, err := repo.FindByID(context.Background(), "missing")
	if !errors.Is(err, domainerrors.ErrVoteNotFound) {
		t.Errorf("expected ErrVoteNotFound, got %v", err)
	}
}

func TestPostgresVoteRepository_ListByMatch_ReturnsAllVotes(t *testing.T) {
	repo := newTestVoteRepository(t)
	ctx := context.Background()
	// Need 2 votes with different voted_id or voter_id to respect UNIQUE.
	_ = repo.Save(ctx, buildTestVote(t, "v-1", "p-voter", "p-voted", 4))

	votes, err := repo.ListByMatch(ctx, "test-match")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(votes) != 1 {
		t.Errorf("expected 1, got %d", len(votes))
	}
}

func TestPostgresVoteRepository_Delete_Removes(t *testing.T) {
	repo := newTestVoteRepository(t)
	ctx := context.Background()
	v := buildTestVote(t, "v-1", "p-voter", "p-voted", 4)
	_ = repo.Save(ctx, v)

	if err := repo.Delete(ctx, "v-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, findErr := repo.FindByID(ctx, "v-1")
	if !errors.Is(findErr, domainerrors.ErrVoteNotFound) {
		t.Errorf("expected ErrVoteNotFound, got %v", findErr)
	}
}
