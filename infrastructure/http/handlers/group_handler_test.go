package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/group"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// ----------------------------------------------------------------------------
// Test helpers — duplicated from application/usecases/group test fakes
// because they live in *_test.go files and are not importable.
// ----------------------------------------------------------------------------

type fakeGroupRepository struct {
	mu     sync.Mutex
	groups map[entities.GroupID]*entities.Group
}

func newFakeGroupRepository() *fakeGroupRepository {
	return &fakeGroupRepository{groups: make(map[entities.GroupID]*entities.Group)}
}

func (r *fakeGroupRepository) Save(_ context.Context, g *entities.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.groups[g.ID()] = g
	return nil
}

func (r *fakeGroupRepository) FindByID(_ context.Context, id entities.GroupID) (*entities.Group, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.groups[id]
	if !ok {
		return nil, domainerrors.ErrGroupNotFound
	}
	return g, nil
}

func (r *fakeGroupRepository) ListByOrganizer(_ context.Context, _ entities.OrganizerID) ([]*entities.Group, error) {
	return nil, nil
}

func (r *fakeGroupRepository) Update(_ context.Context, _ *entities.Group) error {
	return nil
}

func (r *fakeGroupRepository) Delete(_ context.Context, _ entities.GroupID) error {
	return nil
}

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

type fakeIDGenerator struct{ id string }

func (g *fakeIDGenerator) Generate() string { return g.id }

// ----------------------------------------------------------------------------
// Setup helper — builds a fully wired handler with fakes.
// ----------------------------------------------------------------------------

type testHandlerEnv struct {
	router *gin.Engine
	repo   *fakeGroupRepository
}

func newTestHandlerEnv(t *testing.T, fixedID string, fixedTime time.Time) *testHandlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := newFakeGroupRepository()
	createUC := group.NewCreateGroupUseCase(repo, &fakeIDGenerator{id: fixedID}, &fakeClock{now: fixedTime})
	getUC := group.NewGetGroupUseCase(repo)
	handler := NewGroupHandler(createUC, getUC)

	router := gin.New()
	api := router.Group("/api/v1")
	handler.Register(api)

	return &testHandlerEnv{router: router, repo: repo}
}

// ----------------------------------------------------------------------------
// Tests
// ----------------------------------------------------------------------------

func TestGroupHandler_Create_Returns201_WhenRequestIsValid(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2026, 4, 10, 14, 0, 0, 0, time.UTC)
	env := newTestHandlerEnv(t, "group-fixed-id", fixedTime)

	body := `{"name":"Foot du jeudi","organizer_id":"org-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var response map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("expected valid JSON response, got error: %v", err)
	}
	if response["id"] != "group-fixed-id" {
		t.Errorf("expected id 'group-fixed-id', got %v", response["id"])
	}
	if response["name"] != "Foot du jeudi" {
		t.Errorf("expected name 'Foot du jeudi', got %v", response["name"])
	}
	if response["organizer_id"] != "org-1" {
		t.Errorf("expected organizer_id 'org-1', got %v", response["organizer_id"])
	}
}

func TestGroupHandler_Create_Returns400_WhenBodyIsInvalidJSON(t *testing.T) {
	t.Parallel()

	env := newTestHandlerEnv(t, "group-1", time.Now())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(`{not json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestGroupHandler_Create_Returns400_WhenRequiredFieldsAreMissing(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		body string
	}{
		{name: "missing name", body: `{"organizer_id":"org-1"}`},
		{name: "missing organizer_id", body: `{"name":"Foot"}`},
		{name: "empty name", body: `{"name":"","organizer_id":"org-1"}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			env := newTestHandlerEnv(t, "group-1", time.Now())

			req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(testCase.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			env.router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400 for body %q, got %d", testCase.body, rec.Code)
			}
		})
	}
}

func TestGroupHandler_Get_Returns200_WhenGroupExists(t *testing.T) {
	t.Parallel()

	env := newTestHandlerEnv(t, "group-1", time.Now())

	// Pre-seed: create a group via the POST endpoint.
	createBody := `{"name":"Existing","organizer_id":"org-1"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	env.router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup failed: create returned %d", createRec.Code)
	}

	// Now fetch it.
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/groups/group-1", nil)
	getRec := httptest.NewRecorder()
	env.router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body: %s)", getRec.Code, getRec.Body.String())
	}

	var response map[string]any
	_ = json.NewDecoder(getRec.Body).Decode(&response)
	if response["name"] != "Existing" {
		t.Errorf("expected name 'Existing', got %v", response["name"])
	}
}

func TestGroupHandler_Get_Returns404_WhenGroupDoesNotExist(t *testing.T) {
	t.Parallel()

	env := newTestHandlerEnv(t, "group-1", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/nonexistent", nil)
	rec := httptest.NewRecorder()

	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}

	var response map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&response)
	if response["error"] != "group not found" {
		t.Errorf("expected error 'group not found', got %q", response["error"])
	}
}

func TestGroupHandler_Create_PersistsGroupInRepository(t *testing.T) {
	t.Parallel()

	env := newTestHandlerEnv(t, "group-persisted", time.Now())

	body := `{"name":"Persisted","organizer_id":"org-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	env.router.ServeHTTP(rec, req)

	stored, err := env.repo.FindByID(context.Background(), "group-persisted")
	if err != nil {
		t.Fatalf("expected group to be persisted, got error: %v", err)
	}
	if stored.Name() != "Persisted" {
		t.Errorf("expected stored name 'Persisted', got %q", stored.Name())
	}
}
