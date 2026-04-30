//go:build integration

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestBuildRouter_ErrorPaths exercises the error branches every
// handler should produce: 400 on bad input, 401 on missing/invalid
// JWT and bad credentials, 404 on missing entities, 409 on invalid
// state transitions. These paths are deliberately not covered by the
// happy-flow smoke test in flow_integration_test.go; running them
// here lifts the use-case Execute branches that returned errors
// from "untested" to "tested".
func TestBuildRouter_ErrorPaths(t *testing.T) {
	if sharedPool == nil {
		t.Fatal("sharedPool is nil; TestMain did not run setupBootContainer")
	}

	router := buildRouter(sharedPool, nil, "test-jwt-secret-for-error-paths")
	server := httptest.NewServer(router)
	defer server.Close()

	c := newAPIClient(server.URL)

	// Setup: a real organizer + match in Draft so we have a target
	// for transition-error assertions.
	c.postExpect(t, http.StatusCreated, "/api/v1/auth/register", "", map[string]any{
		"email":        "errors@example.com",
		"password":     "errors-test-password",
		"display_name": "Error Tester",
	}, nil)
	var login struct {
		Token     string `json:"token"`
		Organizer struct {
			ID string `json:"id"`
		} `json:"organizer"`
	}
	c.postExpect(t, http.StatusOK, "/api/v1/auth/login", "", map[string]any{
		"email":    "errors@example.com",
		"password": "errors-test-password",
	}, &login)
	c.token = login.Token

	var group struct {
		ID string `json:"id"`
	}
	c.postExpect(t, http.StatusCreated, "/api/v1/groups", c.token, map[string]any{
		"name": "Error Group",
	}, &group)

	var draftMatch struct {
		ID string `json:"id"`
	}
	c.postExpect(t, http.StatusCreated, "/api/v1/matches", c.token, map[string]any{
		"group_id":     group.ID,
		"title":        "Draft Match",
		"venue":        "Stadium",
		"scheduled_at": time.Date(2026, 6, 15, 19, 0, 0, 0, time.UTC),
	}, &draftMatch)

	t.Run("auth: malformed register body returns 400", func(t *testing.T) {
		c.expectStatus(t, http.MethodPost, "/api/v1/auth/register", "",
			rawBody(`{not even json`), http.StatusBadRequest)
	})

	t.Run("auth: register missing email returns 400", func(t *testing.T) {
		c.expectStatus(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
			"password":     "valid-password",
			"display_name": "X",
		}, http.StatusBadRequest)
	})

	t.Run("auth: login wrong password returns 401", func(t *testing.T) {
		c.expectStatus(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
			"email":    "errors@example.com",
			"password": "wrong-password",
		}, http.StatusUnauthorized)
	})

	t.Run("auth: login unknown email returns 401", func(t *testing.T) {
		c.expectStatus(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
			"email":    "nobody-registered@example.com",
			"password": "irrelevant",
		}, http.StatusUnauthorized)
	})

	t.Run("auth: protected endpoint without JWT returns 401", func(t *testing.T) {
		c.expectStatus(t, http.MethodPost, "/api/v1/groups", "", map[string]any{
			"name": "Should Fail",
		}, http.StatusUnauthorized)
	})

	t.Run("auth: protected endpoint with garbage JWT returns 401", func(t *testing.T) {
		c.expectStatus(t, http.MethodPost, "/api/v1/groups", "not-a-real-jwt", map[string]any{
			"name": "Should Fail",
		}, http.StatusUnauthorized)
	})

	t.Run("404: get unknown group returns 404", func(t *testing.T) {
		c.expectStatus(t, http.MethodGet, "/api/v1/groups/does-not-exist", "", nil, http.StatusNotFound)
	})

	t.Run("404: get unknown match returns 404", func(t *testing.T) {
		c.expectStatus(t, http.MethodGet, "/api/v1/matches/does-not-exist", "", nil, http.StatusNotFound)
	})

	t.Run("404: get unknown player returns 404", func(t *testing.T) {
		c.expectStatus(t, http.MethodGet, "/api/v1/players/does-not-exist", "", nil, http.StatusNotFound)
	})

	t.Run("404: get unknown invitation returns 404", func(t *testing.T) {
		c.expectStatus(t, http.MethodGet, "/api/v1/invitations/does-not-exist", "", nil, http.StatusNotFound)
	})

	t.Run("404: get unknown vote returns 404", func(t *testing.T) {
		c.expectStatus(t, http.MethodGet, "/api/v1/votes/does-not-exist", "", nil, http.StatusNotFound)
	})

	t.Run("400: create group missing name returns 400", func(t *testing.T) {
		c.expectStatus(t, http.MethodPost, "/api/v1/groups", c.token, map[string]any{},
			http.StatusBadRequest)
	})

	t.Run("400: create match missing required field returns 400", func(t *testing.T) {
		c.expectStatus(t, http.MethodPost, "/api/v1/matches", c.token, map[string]any{
			"group_id": group.ID,
			"title":    "",
		}, http.StatusBadRequest)
	})

	t.Run("400: cast vote with score out of range returns 400", func(t *testing.T) {
		c.expectStatus(t, http.MethodPost, "/api/v1/votes", "", map[string]any{
			"match_id": draftMatch.ID,
			"voter_id": "p1",
			"voted_id": "p2",
			"score":    99,
		}, http.StatusBadRequest)
	})

	t.Run("400: accept invitation with garbage token returns 4xx", func(t *testing.T) {
		// Token-shape errors and not-found can both surface here. We
		// only assert the response is a 4xx — more precise than that
		// and the test couples to an internal mapping detail.
		resp, _ := c.do(t, http.MethodPost, "/api/v1/invitations/accept", "", map[string]any{
			"token": "definitely-not-a-real-token",
		})
		if resp.StatusCode < 400 || resp.StatusCode >= 500 {
			t.Errorf("expected 4xx for bad accept token, got %d", resp.StatusCode)
		}
	})

	t.Run("409: start match while still in Draft returns 409", func(t *testing.T) {
		// Draft → Start is not a legal transition. The use case
		// returns ErrInvalidTransition which maps to 409 Conflict.
		c.expectStatus(t, http.MethodPost,
			"/api/v1/matches/"+draftMatch.ID+"/start", c.token, nil,
			http.StatusConflict)
	})

	t.Run("409: complete match while still in Draft returns 409", func(t *testing.T) {
		c.expectStatus(t, http.MethodPost,
			"/api/v1/matches/"+draftMatch.ID+"/complete", c.token, nil,
			http.StatusConflict)
	})
}

// rawBody is a marker type: the apiClient sends it as the literal
// request body without JSON-marshalling it, so tests can submit
// intentionally malformed payloads. Recognised by apiClient.do().
type rawBody string

// expectStatus is a thin wrapper around do() that only asserts the
// response status. Used for the many error-path tests that don't
// care about the response body shape.
func (c *apiClient) expectStatus(t *testing.T, method, path, token string, body any, want int) {
	t.Helper()
	resp, raw := c.do(t, method, path, token, body)
	if resp.StatusCode != want {
		t.Errorf("%s %s: got %d, want %d. body=%s", method, path, resp.StatusCode, want, string(raw))
	}
}
