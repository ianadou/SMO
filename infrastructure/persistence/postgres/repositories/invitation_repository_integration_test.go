//go:build integration

package repositories_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/repositories"
)

// newTestInvitationRepository cleans the invitations table and re-seeds a
// match so invitation rows can reference it via FK.
func newTestInvitationRepository(t *testing.T) *repositories.PostgresInvitationRepository {
	t.Helper()
	ctx := context.Background()

	if _, err := sharedPool.Exec(ctx, "DELETE FROM invitations"); err != nil {
		t.Fatalf("clean invitations: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, "DELETE FROM matches"); err != nil {
		t.Fatalf("clean matches: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, `
		INSERT INTO groups (id, organizer_id, name)
		VALUES ('test-group', 'test-org', 'Test Group')
		ON CONFLICT (id) DO NOTHING
	`); err != nil {
		t.Fatalf("re-seed group: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, `
		INSERT INTO matches (id, group_id, title, venue, scheduled_at, status)
		VALUES ('test-match', 'test-group', 'Test Match', 'Venue', NOW() + INTERVAL '1 day', 'open')
	`); err != nil {
		t.Fatalf("seed match: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, `
		INSERT INTO players (id, group_id, name, ranking)
		VALUES ('p-1', 'test-group', 'Test Player', 1000)
		ON CONFLICT (id) DO NOTHING
	`); err != nil {
		t.Fatalf("seed player: %v", err)
	}
	return repositories.NewPostgresInvitationRepository(sharedPool)
}

// seedPlayer adds an extra player so distinct invitations can satisfy
// the player_id FK in the capacity tests.
func seedPlayer(t *testing.T, id string) {
	t.Helper()
	if _, err := sharedPool.Exec(context.Background(), `
		INSERT INTO players (id, group_id, name, ranking)
		VALUES ($1, 'test-group', $2, 1000)
		ON CONFLICT (id) DO NOTHING
	`, id, "Player "+id); err != nil {
		t.Fatalf("seed player %s: %v", id, err)
	}
}

func buildTestInvitation(t *testing.T, id, matchID, hash string) *entities.Invitation {
	t.Helper()
	createdAt := time.Now()
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	inv, err := entities.NewInvitation(
		entities.InvitationID(id), entities.MatchID(matchID), entities.PlayerID("p-1"), hash,
		expiresAt, entities.InvitationResponsePending, nil, createdAt,
	)
	if err != nil {
		t.Fatalf("NewInvitation: %v", err)
	}
	return inv
}

func TestPostgresInvitationRepository_Save_PersistsPendingInvitation(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	inv := buildTestInvitation(t, "inv-1", "test-match", "hash-1")
	if err := repo.Save(ctx, inv); err != nil {
		t.Fatalf("expected Save to succeed: %v", err)
	}

	found, findErr := repo.FindByID(ctx, "inv-1")
	if findErr != nil {
		t.Fatalf("expected to find saved invitation: %v", findErr)
	}
	if found.TokenHash() != "hash-1" {
		t.Errorf("expected hash 'hash-1', got %q", found.TokenHash())
	}
	if found.Response() != entities.InvitationResponsePending {
		t.Errorf("expected pending response on a new invitation, got %q", found.Response())
	}
}

func TestPostgresInvitationRepository_Save_ReturnsErrReferencedEntityNotFound(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	inv := buildTestInvitation(t, "inv-1", "nonexistent-match", "hash")
	err := repo.Save(ctx, inv)
	if !errors.Is(err, domainerrors.ErrReferencedEntityNotFound) {
		t.Errorf("expected ErrReferencedEntityNotFound, got %v", err)
	}
}

func TestPostgresInvitationRepository_FindByID_ReturnsErrInvitationNotFound(t *testing.T) {
	repo := newTestInvitationRepository(t)
	_, err := repo.FindByID(context.Background(), "does-not-exist")
	if !errors.Is(err, domainerrors.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestPostgresInvitationRepository_FindByTokenHash_ReturnsInvitation(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	inv := buildTestInvitation(t, "inv-1", "test-match", "unique-hash-xyz")
	_ = repo.Save(ctx, inv)

	found, err := repo.FindByTokenHash(ctx, "unique-hash-xyz")
	if err != nil {
		t.Fatalf("expected to find by hash, got: %v", err)
	}
	if found.ID() != "inv-1" {
		t.Errorf("expected inv-1, got %q", found.ID())
	}
}

func TestPostgresInvitationRepository_FindByTokenHash_ReturnsErrInvitationNotFound(t *testing.T) {
	repo := newTestInvitationRepository(t)
	_, err := repo.FindByTokenHash(context.Background(), "nonexistent-hash")
	if !errors.Is(err, domainerrors.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestPostgresInvitationRepository_ListByMatch_ReturnsAll(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	for i, hash := range []string{"h1", "h2", "h3"} {
		inv := buildTestInvitation(t, "inv-"+string(rune('1'+i)), "test-match", hash)
		if err := repo.Save(ctx, inv); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	invitations, err := repo.ListByMatch(ctx, "test-match")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(invitations) != 3 {
		t.Errorf("expected 3, got %d", len(invitations))
	}
}

func TestPostgresInvitationRepository_RespondWithCapacityGuard_PersistsResponse(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	inv := buildTestInvitation(t, "inv-1", "test-match", "hash")
	_ = repo.Save(ctx, inv)

	if err := inv.Respond(entities.InvitationResponseYes, time.Now(), false); err != nil {
		t.Fatalf("entity Respond: %v", err)
	}
	if err := repo.RespondWithCapacityGuard(ctx, inv, entities.MaxParticipantsPerMatch); err != nil {
		t.Fatalf("repo.RespondWithCapacityGuard: %v", err)
	}

	found, _ := repo.FindByID(ctx, "inv-1")
	if !found.IsConfirmed() {
		t.Error("expected invitation to be confirmed after responding yes")
	}
	if found.RespondedAt() == nil {
		t.Error("expected responded_at to be set")
	}
}

func TestPostgresInvitationRepository_RespondWithCapacityGuard_AllowsChangingYesToNo(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	inv := buildTestInvitation(t, "inv-1", "test-match", "hash")
	_ = repo.Save(ctx, inv)
	_ = inv.Respond(entities.InvitationResponseYes, time.Now(), false)
	if err := repo.RespondWithCapacityGuard(ctx, inv, entities.MaxParticipantsPerMatch); err != nil {
		t.Fatalf("first respond: %v", err)
	}

	_ = inv.Respond(entities.InvitationResponseNo, time.Now(), false)
	if err := repo.RespondWithCapacityGuard(ctx, inv, entities.MaxParticipantsPerMatch); err != nil {
		t.Fatalf("change to no: %v", err)
	}

	found, _ := repo.FindByID(ctx, "inv-1")
	if found.Response() != entities.InvitationResponseNo {
		t.Errorf("expected response 'no', got %q", found.Response())
	}
}

func TestPostgresInvitationRepository_Delete_RemovesInvitation(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	inv := buildTestInvitation(t, "inv-1", "test-match", "hash")
	_ = repo.Save(ctx, inv)
	if err := repo.Delete(ctx, "inv-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, findErr := repo.FindByID(ctx, "inv-1")
	if !errors.Is(findErr, domainerrors.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound after delete, got %v", findErr)
	}
}

func TestPostgresInvitationRepository_CountConfirmedByMatch_OnlyCountsYes(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	// Two confirmed (yes) and one declined (no) invitation for the match.
	for i, answer := range []entities.InvitationResponse{
		entities.InvitationResponseYes,
		entities.InvitationResponseYes,
		entities.InvitationResponseNo,
	} {
		seedPlayer(t, fmt.Sprintf("p-count-%d", i))
		createdAt := time.Now()
		inv, err := entities.NewInvitation(
			entities.InvitationID(fmt.Sprintf("inv-%d", i)), "test-match",
			entities.PlayerID(fmt.Sprintf("p-count-%d", i)), fmt.Sprintf("h-%d", i),
			createdAt.Add(48*time.Hour), entities.InvitationResponsePending, nil, createdAt,
		)
		if err != nil {
			t.Fatalf("build invitation: %v", err)
		}
		if err := repo.Save(ctx, inv); err != nil {
			t.Fatalf("Save: %v", err)
		}
		_ = inv.Respond(answer, time.Now(), false)
		if err := repo.RespondWithCapacityGuard(ctx, inv, entities.MaxParticipantsPerMatch); err != nil {
			t.Fatalf("respond: %v", err)
		}
	}

	count, err := repo.CountConfirmedByMatch(ctx, "test-match")
	if err != nil {
		t.Fatalf("CountConfirmedByMatch: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 confirmed (declined excluded), got %d", count)
	}
}

func TestPostgresInvitationRepository_ListConfirmedParticipants_JoinsPlayerName(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	inv := buildTestInvitation(t, "inv-1", "test-match", "hash-1")
	if err := repo.Save(ctx, inv); err != nil {
		t.Fatalf("Save: %v", err)
	}
	_ = inv.Respond(entities.InvitationResponseYes, time.Now(), false)
	if err := repo.RespondWithCapacityGuard(ctx, inv, entities.MaxParticipantsPerMatch); err != nil {
		t.Fatalf("respond: %v", err)
	}

	participants, err := repo.ListConfirmedParticipants(ctx, "test-match")
	if err != nil {
		t.Fatalf("ListConfirmedParticipants: %v", err)
	}
	if len(participants) != 1 {
		t.Fatalf("expected 1 participant, got %d", len(participants))
	}
	if participants[0].PlayerID != "p-1" {
		t.Errorf("expected player_id 'p-1', got %q", participants[0].PlayerID)
	}
	if participants[0].PlayerName != "Test Player" {
		t.Errorf("expected player_name 'Test Player' from JOIN, got %q", participants[0].PlayerName)
	}
	if participants[0].ConfirmedAt.IsZero() {
		t.Error("expected ConfirmedAt to be set")
	}
}

func TestPostgresInvitationRepository_ListConfirmedParticipants_ExcludesPendingAndDeclined(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	pending := buildTestInvitation(t, "inv-pending", "test-match", "hash-p")
	if err := repo.Save(ctx, pending); err != nil {
		t.Fatalf("Save pending: %v", err)
	}

	seedPlayer(t, "p-declined")
	createdAt := time.Now()
	declined, _ := entities.NewInvitation(
		"inv-declined", "test-match", "p-declined", "hash-d",
		createdAt.Add(48*time.Hour), entities.InvitationResponsePending, nil, createdAt,
	)
	if err := repo.Save(ctx, declined); err != nil {
		t.Fatalf("Save declined: %v", err)
	}
	_ = declined.Respond(entities.InvitationResponseNo, time.Now(), false)
	if err := repo.RespondWithCapacityGuard(ctx, declined, entities.MaxParticipantsPerMatch); err != nil {
		t.Fatalf("respond no: %v", err)
	}

	participants, err := repo.ListConfirmedParticipants(ctx, "test-match")
	if err != nil {
		t.Fatalf("ListConfirmedParticipants: %v", err)
	}
	if len(participants) != 0 {
		t.Errorf("expected 0 participants (pending+declined only), got %d", len(participants))
	}
}

// TestPostgresInvitationRepository_MigrationBackfill verifies the 006
// migration backfilled legacy used_at rows to response='yes' with
// responded_at carrying the original timestamp. The migration already
// ran during container setup, so we reproduce its precondition (a row
// with used_at set) by writing the row and the legacy column directly,
// then re-running the idempotent backfill statement and asserting the
// result.
func TestPostgresInvitationRepository_MigrationBackfill(t *testing.T) {
	newTestInvitationRepository(t)
	ctx := context.Background()

	usedAt := time.Date(2026, 6, 1, 18, 30, 0, 0, time.UTC)
	if _, err := sharedPool.Exec(ctx, `
		INSERT INTO invitations
			(id, match_id, player_id, token_hash, expires_at, used_at, response, responded_at, created_at)
		VALUES
			('legacy-1', 'test-match', 'p-1', 'legacy-hash',
			 NOW() + INTERVAL '2 days', $1, 'pending', NULL, NOW())
	`, usedAt); err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	// Idempotent re-application of the migration's backfill statement.
	if _, err := sharedPool.Exec(ctx, `
		UPDATE invitations
		SET response = 'yes', responded_at = used_at
		WHERE used_at IS NOT NULL AND response = 'pending'
	`); err != nil {
		t.Fatalf("backfill: %v", err)
	}

	var response string
	var respondedAt time.Time
	if err := sharedPool.QueryRow(ctx, `
		SELECT response, responded_at FROM invitations WHERE id = 'legacy-1'
	`).Scan(&response, &respondedAt); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if response != "yes" {
		t.Errorf("expected backfilled response='yes', got %q", response)
	}
	if !respondedAt.Equal(usedAt) {
		t.Errorf("expected responded_at=used_at (%v), got %v", usedAt, respondedAt)
	}
}

// TestPostgresInvitationRepository_RespondWithCapacityGuard_IsConcurrencySafe
// proves the FOR UPDATE serialization: with exactly one free slot,
// firing N concurrent "yes" confirmations must let exactly one through
// and reject the rest with ErrMatchFull — never overshoot the cap.
func TestPostgresInvitationRepository_RespondWithCapacityGuard_IsConcurrencySafe(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	// Fill the match to exactly MaxParticipantsPerMatch-1 confirmed.
	for i := 0; i < entities.MaxParticipantsPerMatch-1; i++ {
		seedPlayer(t, fmt.Sprintf("p-seed-%d", i))
		createdAt := time.Now()
		inv, err := entities.NewInvitation(
			entities.InvitationID(fmt.Sprintf("inv-seed-%d", i)), "test-match",
			entities.PlayerID(fmt.Sprintf("p-seed-%d", i)), fmt.Sprintf("seed-h-%d", i),
			createdAt.Add(48*time.Hour), entities.InvitationResponsePending, nil, createdAt,
		)
		if err != nil {
			t.Fatalf("build seed invitation: %v", err)
		}
		if err := repo.Save(ctx, inv); err != nil {
			t.Fatalf("save seed: %v", err)
		}
		_ = inv.Respond(entities.InvitationResponseYes, time.Now(), false)
		if err := repo.RespondWithCapacityGuard(ctx, inv, entities.MaxParticipantsPerMatch); err != nil {
			t.Fatalf("seed respond: %v", err)
		}
	}

	// Create N pending invitations all racing for the single free slot.
	const contenders = 6
	pendings := make([]*entities.Invitation, contenders)
	for i := 0; i < contenders; i++ {
		seedPlayer(t, fmt.Sprintf("p-race-%d", i))
		createdAt := time.Now()
		inv, err := entities.NewInvitation(
			entities.InvitationID(fmt.Sprintf("inv-race-%d", i)), "test-match",
			entities.PlayerID(fmt.Sprintf("p-race-%d", i)), fmt.Sprintf("race-h-%d", i),
			createdAt.Add(48*time.Hour), entities.InvitationResponsePending, nil, createdAt,
		)
		if err != nil {
			t.Fatalf("build contender: %v", err)
		}
		if err := repo.Save(ctx, inv); err != nil {
			t.Fatalf("save contender: %v", err)
		}
		pendings[i] = inv
	}

	var wg sync.WaitGroup
	results := make([]error, contenders)
	for i := 0; i < contenders; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			inv := pendings[idx]
			_ = inv.Respond(entities.InvitationResponseYes, time.Now(), false)
			results[idx] = repo.RespondWithCapacityGuard(ctx, inv, entities.MaxParticipantsPerMatch)
		}(i)
	}
	wg.Wait()

	successes, full := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domainerrors.ErrMatchFull):
			full++
		default:
			t.Fatalf("unexpected error from concurrent respond: %v", err)
		}
	}
	if successes != 1 {
		t.Errorf("expected exactly 1 successful confirmation, got %d", successes)
	}
	if full != contenders-1 {
		t.Errorf("expected %d ErrMatchFull rejections, got %d", contenders-1, full)
	}

	confirmed, err := repo.CountConfirmedByMatch(ctx, "test-match")
	if err != nil {
		t.Fatalf("final count: %v", err)
	}
	if confirmed != entities.MaxParticipantsPerMatch {
		t.Errorf("expected exactly %d confirmed after the race, got %d",
			entities.MaxParticipantsPerMatch, confirmed)
	}
}
