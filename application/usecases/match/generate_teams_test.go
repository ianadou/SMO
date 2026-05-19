package match

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// --- fakes specific to GenerateTeamsUseCase --------------------------------

type fakeInvitationRepoForGenerate struct {
	mu           sync.Mutex
	participants map[entities.MatchID][]entities.MatchParticipant
}

func newFakeInvitationRepoForGenerate() *fakeInvitationRepoForGenerate {
	return &fakeInvitationRepoForGenerate{
		participants: make(map[entities.MatchID][]entities.MatchParticipant),
	}
}

func (r *fakeInvitationRepoForGenerate) ListConfirmedParticipants(
	_ context.Context, matchID entities.MatchID,
) ([]entities.MatchParticipant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]entities.MatchParticipant(nil), r.participants[matchID]...), nil
}

func (r *fakeInvitationRepoForGenerate) Save(context.Context, *entities.Invitation) error {
	panic("not used")
}

func (r *fakeInvitationRepoForGenerate) FindByID(
	context.Context, entities.InvitationID,
) (*entities.Invitation, error) {
	panic("not used")
}

func (r *fakeInvitationRepoForGenerate) FindByTokenHash(
	context.Context, string,
) (*entities.Invitation, error) {
	panic("not used")
}

func (r *fakeInvitationRepoForGenerate) ListByMatch(
	context.Context, entities.MatchID,
) ([]*entities.Invitation, error) {
	panic("not used")
}

func (r *fakeInvitationRepoForGenerate) CountConfirmedByMatch(
	context.Context, entities.MatchID,
) (int, error) {
	panic("not used")
}

func (r *fakeInvitationRepoForGenerate) RespondWithCapacityGuard(
	context.Context, *entities.Invitation, int,
) error {
	panic("not used")
}

func (r *fakeInvitationRepoForGenerate) Delete(context.Context, entities.InvitationID) error {
	panic("not used")
}

type fakePlayerRepoForGenerate struct {
	mu      sync.Mutex
	players map[entities.PlayerID]*entities.Player
}

func newFakePlayerRepoForGenerate() *fakePlayerRepoForGenerate {
	return &fakePlayerRepoForGenerate{
		players: make(map[entities.PlayerID]*entities.Player),
	}
}

func (r *fakePlayerRepoForGenerate) Save(_ context.Context, p *entities.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.players[p.ID()] = p
	return nil
}

func (r *fakePlayerRepoForGenerate) FindByID(
	_ context.Context, id entities.PlayerID,
) (*entities.Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.players[id]
	if !ok {
		return nil, domainerrors.ErrPlayerNotFound
	}
	return p, nil
}

func (r *fakePlayerRepoForGenerate) ListByGroup(
	context.Context, entities.GroupID,
) ([]*entities.Player, error) {
	panic("not used")
}

func (r *fakePlayerRepoForGenerate) UpdateRanking(context.Context, *entities.Player) error {
	panic("not used")
}

func (r *fakePlayerRepoForGenerate) Delete(context.Context, entities.PlayerID) error {
	panic("not used")
}

// --- helpers ----------------------------------------------------------------

const generateTestMatchID entities.MatchID = "m1"

func newGenerateUseCase(t *testing.T) (
	*GenerateTeamsUseCase,
	*fakeMatchRepository,
	*fakeInvitationRepoForGenerate,
	*fakePlayerRepoForGenerate,
) {
	t.Helper()
	matchRepo := newFakeMatchRepository()
	invRepo := newFakeInvitationRepoForGenerate()
	playerRepo := newFakePlayerRepoForGenerate()
	clock := newFakeClock(time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC))
	uc := NewGenerateTeamsUseCase(matchRepo, invRepo, playerRepo, clock)
	return uc, matchRepo, invRepo, playerRepo
}

func seedOpenMatch(t *testing.T, repo *fakeMatchRepository) *entities.Match {
	t.Helper()
	m, err := entities.NewMatch(generateTestMatchID, "g-1", "Test", "Venue",
		time.Now().Add(time.Hour), time.Now())
	if err != nil {
		t.Fatalf("seed: NewMatch: %v", err)
	}
	if openErr := m.Open(); openErr != nil {
		t.Fatalf("seed: Open: %v", openErr)
	}
	if saveErr := repo.Save(context.Background(), m); saveErr != nil {
		t.Fatalf("seed: Save: %v", saveErr)
	}
	return m
}

func seedFourConfirmedPlayers(
	t *testing.T,
	invRepo *fakeInvitationRepoForGenerate,
	playerRepo *fakePlayerRepoForGenerate,
) {
	t.Helper()
	ctx := context.Background()
	specs := []struct {
		id      entities.PlayerID
		name    string
		ranking int
	}{
		{"p1", "Alice", 1400},
		{"p2", "Bob", 1300},
		{"p3", "Carol", 1200},
		{"p4", "Dave", 1100},
	}
	for _, s := range specs {
		player, err := entities.NewPlayer(s.id, "g-1", s.name, s.ranking)
		if err != nil {
			t.Fatalf("seed: NewPlayer %q: %v", s.id, err)
		}
		if saveErr := playerRepo.Save(ctx, player); saveErr != nil {
			t.Fatalf("seed: Save player %q: %v", s.id, saveErr)
		}
		invRepo.participants[generateTestMatchID] = append(invRepo.participants[generateTestMatchID],
			entities.MatchParticipant{
				PlayerID:    s.id,
				PlayerName:  s.name,
				ConfirmedAt: time.Now(),
			})
	}
}

// --- tests ------------------------------------------------------------------

func TestGenerateTeams_AssignsAllConfirmedPlayers_WithRanking(t *testing.T) {
	t.Parallel()
	uc, matchRepo, invRepo, playerRepo := newGenerateUseCase(t)
	ctx := context.Background()

	seedOpenMatch(t, matchRepo)
	seedFourConfirmedPlayers(t, invRepo, playerRepo)

	result, err := uc.Execute(ctx, generateTestMatchID, "ranking")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !result.HasTeams() {
		t.Errorf("expected match to have teams")
	}
	total := len(result.TeamA()) + len(result.TeamB())
	if total != 4 {
		t.Errorf("expected 4 players assigned, got %d", total)
	}

	if matchRepo.replaceTeamsCalls != 1 {
		t.Fatalf("expected ReplaceTeams to be called exactly once, got %d", matchRepo.replaceTeamsCalls)
	}

	persisted, ferr := matchRepo.FindByID(ctx, generateTestMatchID)
	if ferr != nil {
		t.Fatalf("expected persisted match, got error: %v", ferr)
	}
	if !persisted.HasTeams() {
		t.Errorf("expected persisted match to have teams")
	}
	persistedTotal := len(persisted.TeamA()) + len(persisted.TeamB())
	if persistedTotal != 4 {
		t.Errorf("expected 4 persisted players, got %d", persistedTotal)
	}
}

func TestGenerateTeams_Rejected_WhenMatchNotOpen(t *testing.T) {
	t.Parallel()
	uc, matchRepo, invRepo, playerRepo := newGenerateUseCase(t)
	ctx := context.Background()

	m, _ := entities.NewMatch(generateTestMatchID, "g-1", "Test", "Venue",
		time.Now().Add(time.Hour), time.Now())
	_ = matchRepo.Save(ctx, m)
	seedFourConfirmedPlayers(t, invRepo, playerRepo)

	_, err := uc.Execute(ctx, generateTestMatchID, "ranking")

	if !errors.Is(err, domainerrors.ErrTeamsNotEditable) {
		t.Errorf("expected ErrTeamsNotEditable, got %v", err)
	}
}

func TestGenerateTeams_Rejected_WhenUnknownStrategy(t *testing.T) {
	t.Parallel()
	uc, matchRepo, invRepo, playerRepo := newGenerateUseCase(t)
	ctx := context.Background()

	seedOpenMatch(t, matchRepo)
	seedFourConfirmedPlayers(t, invRepo, playerRepo)

	_, err := uc.Execute(ctx, generateTestMatchID, "bogus")

	if !errors.Is(err, domainerrors.ErrInvalidParameter) {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestGenerateTeams_Random_IsDeterministic_WithFixedClockSeed(t *testing.T) {
	t.Parallel()
	uc, matchRepo, invRepo, playerRepo := newGenerateUseCase(t)
	ctx := context.Background()

	seedOpenMatch(t, matchRepo)
	seedFourConfirmedPlayers(t, invRepo, playerRepo)

	first, err := uc.Execute(ctx, generateTestMatchID, "random")
	if err != nil {
		t.Fatalf("first run: expected no error, got: %v", err)
	}
	firstA := append([]entities.PlayerID(nil), first.TeamA()...)
	firstB := append([]entities.PlayerID(nil), first.TeamB()...)

	seedOpenMatch(t, matchRepo)

	second, err := uc.Execute(ctx, generateTestMatchID, "random")
	if err != nil {
		t.Fatalf("second run: expected no error, got: %v", err)
	}

	if !equalPlayerIDs(firstA, second.TeamA()) {
		t.Errorf("teamA not deterministic: %v vs %v", firstA, second.TeamA())
	}
	if !equalPlayerIDs(firstB, second.TeamB()) {
		t.Errorf("teamB not deterministic: %v vs %v", firstB, second.TeamB())
	}
}

func equalPlayerIDs(a, b []entities.PlayerID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
