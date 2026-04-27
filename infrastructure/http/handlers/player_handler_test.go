package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/player"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/http/handlers"
)

// --- fakes ---------------------------------------------------------------

type fakePlayerRepo struct {
	mu      sync.Mutex
	players map[entities.PlayerID]*entities.Player
}

func newFakePlayerRepo() *fakePlayerRepo {
	return &fakePlayerRepo{players: make(map[entities.PlayerID]*entities.Player)}
}

func (r *fakePlayerRepo) Save(_ context.Context, p *entities.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.players[p.ID()] = p
	return nil
}

func (r *fakePlayerRepo) FindByID(_ context.Context, id entities.PlayerID) (*entities.Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.players[id]
	if !ok {
		return nil, domainerrors.ErrPlayerNotFound
	}
	return p, nil
}

func (r *fakePlayerRepo) ListByGroup(_ context.Context, groupID entities.GroupID) ([]*entities.Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*entities.Player, 0)
	for _, p := range r.players {
		if p.GroupID() == groupID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *fakePlayerRepo) UpdateRanking(_ context.Context, p *entities.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.players[p.ID()]; !ok {
		return domainerrors.ErrPlayerNotFound
	}
	r.players[p.ID()] = p
	return nil
}

func (r *fakePlayerRepo) Delete(_ context.Context, id entities.PlayerID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.players, id)
	return nil
}

type playerFixedIDGen struct{ id string }

func (g *playerFixedIDGen) Generate() string { return g.id }

// --- helpers -------------------------------------------------------------

func buildPlayerTestRouter(t *testing.T) (*gin.Engine, *fakePlayerRepo) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := newFakePlayerRepo()
	idGen := &playerFixedIDGen{id: "generated-id"}

	handler := handlers.NewPlayerHandler(
		player.NewCreatePlayerUseCase(repo, idGen),
		player.NewGetPlayerUseCase(repo),
		player.NewListPlayersByGroupUseCase(repo),
		player.NewUpdatePlayerRankingUseCase(repo),
	)

	router := gin.New()
	api := router.Group("/api/v1")
	handler.Register(api)
	return router, repo
}

func seedPlayer(t *testing.T, repo *fakePlayerRepo) *entities.Player {
	t.Helper()
	p, err := entities.NewPlayer("player-1", "group-1", "Alice", 1000)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	_ = repo.Save(context.Background(), p)
	return p
}

// --- tests ---------------------------------------------------------------

func TestPlayerHandler_Create_Returns201_WithDefaultRanking(t *testing.T) {
	router, _ := buildPlayerTestRouter(t)

	body := `{"group_id":"group-1","name":"Alice"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/players", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["id"] != "generated-id" {
		t.Errorf("expected id 'generated-id', got %v", resp["id"])
	}
	// Default ranking is 1000; JSON numbers decode to float64.
	if resp["ranking"].(float64) != 1000 {
		t.Errorf("expected ranking 1000, got %v", resp["ranking"])
	}
}

func TestPlayerHandler_Create_Returns400_WhenBodyMissingFields(t *testing.T) {
	router, _ := buildPlayerTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/players", bytes.NewBufferString(`{"name":"Alice"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestPlayerHandler_Get_Returns200_WhenExists(t *testing.T) {
	router, repo := buildPlayerTestRouter(t)
	seedPlayer(t, repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/players/player-1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestPlayerHandler_Get_Returns404_WhenMissing(t *testing.T) {
	router, _ := buildPlayerTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/players/nonexistent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestPlayerHandler_ListByGroup_Returns200_WithArray(t *testing.T) {
	router, repo := buildPlayerTestRouter(t)
	seedPlayer(t, repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/group-1/players", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var list []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Errorf("expected 1 player, got %d", len(list))
	}
}

func TestPlayerHandler_UpdateRanking_Returns200_AndPersistsNewValue(t *testing.T) {
	router, repo := buildPlayerTestRouter(t)
	seedPlayer(t, repo)

	body := `{"ranking":1500}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/players/player-1/ranking", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["ranking"].(float64) != 1500 {
		t.Errorf("expected ranking 1500, got %v", resp["ranking"])
	}
}

func TestPlayerHandler_UpdateRanking_Returns404_WhenPlayerMissing(t *testing.T) {
	router, _ := buildPlayerTestRouter(t)

	body := `{"ranking":1500}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/players/nonexistent/ranking", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
