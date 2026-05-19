//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestBuildRouter_FullOrganizerFlow exercises every major endpoint of
// the API in one end-to-end smoke run. The intent is coverage breadth:
// hitting each handler+use-case Execute path catches wiring bugs that
// unit tests cannot see (route mismatch, middleware order, missing
// dependency, JSON contract drift).
//
// One large test rather than many small tests because the testcontainer
// boot is the expensive part; a single sharedPool reused across all
// HTTP calls keeps total CI time reasonable.
func TestBuildRouter_FullOrganizerFlow(t *testing.T) {
	if sharedPool == nil {
		t.Fatal("sharedPool is nil; TestMain did not run setupBootContainer")
	}

	router := buildRouter(sharedPool, nil, "test-jwt-secret-for-flow-smoke")
	server := httptest.NewServer(router)
	defer server.Close()

	c := newAPIClient(server.URL)

	// 1. Register organizer + login.
	c.postExpect(t, http.StatusCreated, "/api/v1/auth/register", "", map[string]any{
		"email":        "flow@example.com",
		"password":     "flow-test-password",
		"display_name": "Flow Tester",
	}, nil)
	var login struct {
		Token     string `json:"token"`
		Organizer struct {
			ID string `json:"id"`
		} `json:"organizer"`
	}
	c.postExpect(t, http.StatusOK, "/api/v1/auth/login", "", map[string]any{
		"email":    "flow@example.com",
		"password": "flow-test-password",
	}, &login)
	if login.Token == "" || login.Organizer.ID == "" {
		t.Fatalf("login response missing token or organizer id: %+v", login)
	}
	c.token = login.Token

	// 2. Create + read group.
	var group struct {
		ID string `json:"id"`
	}
	c.postExpect(t, http.StatusCreated, "/api/v1/groups", c.token, map[string]any{
		"name": "Flow Group",
	}, &group)
	c.getExpect(t, http.StatusOK, "/api/v1/groups/"+group.ID, "", nil)

	// 3. Create players (need two so a vote has a sensible voter/voted pair).
	var p1, p2 struct {
		ID string `json:"id"`
	}
	c.postExpect(t, http.StatusCreated, "/api/v1/players", c.token, map[string]any{
		"group_id": group.ID,
		"name":     "Alice",
	}, &p1)
	c.postExpect(t, http.StatusCreated, "/api/v1/players", c.token, map[string]any{
		"group_id": group.ID,
		"name":     "Bob",
	}, &p2)
	c.getExpect(t, http.StatusOK, "/api/v1/players/"+p1.ID, "", nil)
	c.getExpect(t, http.StatusOK, "/api/v1/groups/"+group.ID+"/players", "", nil)

	// 4. Create match in the group, exercise every transition.
	var match struct {
		ID string `json:"id"`
	}
	c.postExpect(t, http.StatusCreated, "/api/v1/matches", c.token, map[string]any{
		"group_id":     group.ID,
		"title":        "Flow Match",
		"venue":        "Stadium A",
		"scheduled_at": time.Date(2026, 6, 1, 19, 0, 0, 0, time.UTC),
	}, &match)
	c.getExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID, "", nil)
	c.getExpect(t, http.StatusOK, "/api/v1/groups/"+group.ID+"/matches", "", nil)

	c.postExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID+"/open", c.token, nil, nil)

	// 5. Invitations: organizer creates one, a "player" confirms
	// attendance via the public token endpoint.
	var inv struct {
		ID         string `json:"id"`
		PlainToken string `json:"plain_token"`
	}
	c.postExpect(t, http.StatusCreated, "/api/v1/invitations", c.token, map[string]any{
		"match_id":  match.ID,
		"player_id": p1.ID,
	}, &inv)
	if inv.PlainToken == "" {
		t.Fatalf("invitation response missing plain_token: %+v", inv)
	}
	c.getExpect(t, http.StatusOK, "/api/v1/invitations/"+inv.ID, "", nil)
	c.getExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID+"/invitations", c.token, nil)

	// 5b. Public context endpoint: the player resolves everything the
	// invitation page needs from the token alone (no auth).
	var invCtx struct {
		OrganizerName  string `json:"organizer_name"`
		GroupName      string `json:"group_name"`
		Capacity       string `json:"capacity"`
		ConfirmedCount int    `json:"confirmed_count"`
		Response       string `json:"response"`
		State          string `json:"state"`
	}
	c.postExpect(t, http.StatusOK, "/api/v1/invitations/context", "", map[string]any{
		"token": inv.PlainToken,
	}, &invCtx)
	if invCtx.OrganizerName != "Flow Tester" || invCtx.GroupName != "Flow Group" ||
		invCtx.Capacity != "10 (5v5)" || invCtx.State != "respondable" ||
		invCtx.Response != "pending" || invCtx.ConfirmedCount != 0 {
		t.Fatalf("unexpected invitation context: %+v", invCtx)
	}
	c.postExpect(t, http.StatusNotFound, "/api/v1/invitations/context", "", map[string]any{
		"token": "unknown-token",
	}, nil)

	c.postExpect(t, http.StatusOK, "/api/v1/invitations/respond", "", map[string]any{
		"token":  inv.PlainToken,
		"answer": "yes",
	}, nil)

	// 5c. A second player confirms too, so the match has a roster a
	// strategy can split into two non-empty teams.
	var inv2 struct {
		PlainToken string `json:"plain_token"`
	}
	c.postExpect(t, http.StatusCreated, "/api/v1/invitations", c.token, map[string]any{
		"match_id": match.ID, "player_id": p2.ID,
	}, &inv2)
	c.postExpect(t, http.StatusOK, "/api/v1/invitations/respond", "", map[string]any{
		"token": inv2.PlainToken, "answer": "yes",
	}, nil)

	// 6. Teams must exist before teams_ready; generate then transition.
	c.postExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID+"/teams/generate", c.token, map[string]any{
		"strategy": "random",
	}, nil)
	c.postExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID+"/teams-ready", c.token, nil, nil)
	c.postExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID+"/start", c.token, nil, nil)
	c.postExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID+"/complete", c.token, nil, nil)

	// 7. Cast + read votes (only allowed once the match is completed).
	var vote struct {
		ID string `json:"id"`
	}
	c.postExpect(t, http.StatusCreated, "/api/v1/votes", "", map[string]any{
		"match_id": match.ID,
		"voter_id": string(p1.ID),
		"voted_id": string(p2.ID),
		"score":    5,
	}, &vote)
	c.getExpect(t, http.StatusOK, "/api/v1/votes/"+vote.ID, "", nil)
	c.getExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID+"/votes", "", nil)

	// 8. Update a player's ranking explicitly (not the finalize path).
	c.patchExpect(t, http.StatusOK, "/api/v1/players/"+p1.ID+"/ranking", c.token, map[string]any{
		"ranking": 100,
	}, nil)

	// 9. Finalize: closes the match, recomputes rankings from votes.
	c.postExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID+"/finalize", c.token, nil, nil)
}

// TestBuildRouter_TeamAssignmentFlow exercises the MD1 team endpoints
// end-to-end against a real Postgres: a match is opened with four
// confirmed participants, teams cannot be marked ready until assigned,
// generation by ranking persists a balanced split, GET resolves names,
// an imbalanced manual partition is rejected, a valid one is accepted,
// and only then is teams_ready allowed.
func TestBuildRouter_TeamAssignmentFlow(t *testing.T) {
	if sharedPool == nil {
		t.Fatal("sharedPool is nil; TestMain did not run setupBootContainer")
	}

	router := buildRouter(sharedPool, nil, "test-jwt-secret-for-teams-flow")
	server := httptest.NewServer(router)
	defer server.Close()

	c := newAPIClient(server.URL)

	c.postExpect(t, http.StatusCreated, "/api/v1/auth/register", "", map[string]any{
		"email": "teams@example.com", "password": "teams-test-password", "display_name": "Teams Tester",
	}, nil)
	var login struct {
		Token string `json:"token"`
	}
	c.postExpect(t, http.StatusOK, "/api/v1/auth/login", "", map[string]any{
		"email": "teams@example.com", "password": "teams-test-password",
	}, &login)
	c.token = login.Token

	var group struct {
		ID string `json:"id"`
	}
	c.postExpect(t, http.StatusCreated, "/api/v1/groups", c.token, map[string]any{
		"name": "Teams Group",
	}, &group)

	names := []string{"Alex L.", "Inès R.", "Théo B.", "Marc R."}
	rankings := []int{40, 30, 20, 10}
	playerIDs := make([]string, 0, len(names))
	for i, name := range names {
		var p struct {
			ID string `json:"id"`
		}
		c.postExpect(t, http.StatusCreated, "/api/v1/players", c.token, map[string]any{
			"group_id": group.ID, "name": name,
		}, &p)
		c.patchExpect(t, http.StatusOK, "/api/v1/players/"+p.ID+"/ranking", c.token, map[string]any{
			"ranking": rankings[i],
		}, nil)
		playerIDs = append(playerIDs, p.ID)
	}

	var match struct {
		ID string `json:"id"`
	}
	c.postExpect(t, http.StatusCreated, "/api/v1/matches", c.token, map[string]any{
		"group_id":     group.ID,
		"title":        "Teams Match",
		"venue":        "Stadium B",
		"scheduled_at": time.Date(2026, 7, 1, 19, 0, 0, 0, time.UTC),
	}, &match)
	c.postExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID+"/open", c.token, nil, nil)

	for _, pid := range playerIDs {
		var inv struct {
			PlainToken string `json:"plain_token"`
		}
		c.postExpect(t, http.StatusCreated, "/api/v1/invitations", c.token, map[string]any{
			"match_id": match.ID, "player_id": pid,
		}, &inv)
		c.postExpect(t, http.StatusOK, "/api/v1/invitations/respond", "", map[string]any{
			"token": inv.PlainToken, "answer": "yes",
		}, nil)
	}

	// teams_ready is refused until a composition exists.
	c.postExpect(t, http.StatusConflict, "/api/v1/matches/"+match.ID+"/teams-ready", c.token, nil, nil)

	var generated []map[string]any
	c.postExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID+"/teams/generate", c.token, map[string]any{
		"strategy": "ranking",
	}, &generated)
	if len(generated) != len(playerIDs) {
		t.Fatalf("generate: expected %d members, got %d: %+v", len(playerIDs), len(generated), generated)
	}

	var fetched []map[string]any
	c.getExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID+"/teams", c.token, &fetched)
	foundName := false
	for _, m := range fetched {
		if m["player_name"] == "Alex L." {
			foundName = true
		}
	}
	if !foundName {
		t.Fatalf("GET teams did not resolve player names: %+v", fetched)
	}

	// Imbalanced manual partition (3 vs 1) is a domain violation → 400.
	c.expect(t, http.MethodPut, "/api/v1/matches/"+match.ID+"/teams", c.token, map[string]any{
		"team_a": playerIDs[:3], "team_b": playerIDs[3:],
	}, nil, http.StatusBadRequest)

	// Valid 2v2 partition is accepted.
	c.expect(t, http.MethodPut, "/api/v1/matches/"+match.ID+"/teams", c.token, map[string]any{
		"team_a": playerIDs[:2], "team_b": playerIDs[2:],
	}, nil, http.StatusOK)

	// With teams assigned, teams_ready now succeeds.
	c.postExpect(t, http.StatusOK, "/api/v1/matches/"+match.ID+"/teams-ready", c.token, nil, nil)
}

// apiClient is a tiny HTTP wrapper that handles JSON marshalling,
// JWT bearer headers, and status assertions in one place.
type apiClient struct {
	baseURL string
	token   string
}

func newAPIClient(baseURL string) *apiClient {
	return &apiClient{baseURL: baseURL}
}

func (c *apiClient) do(t *testing.T, method, path, token string, body any) (*http.Response, []byte) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		// rawBody (defined in flow_errors_integration_test.go) is sent
		// verbatim so tests can submit intentionally malformed JSON.
		// Anything else is JSON-marshalled.
		if raw, ok := body.(rawBody); ok {
			reader = bytes.NewReader([]byte(raw))
		} else {
			marshalled, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("marshal %s %s: %v", method, path, err)
			}
			reader = bytes.NewReader(marshalled)
		}
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		t.Fatalf("new request %s %s: %v", method, path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do %s %s: %v", method, path, err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return resp, respBody
}

func (c *apiClient) expect(t *testing.T, method, path, token string, body, into any, want int) {
	t.Helper()
	resp, raw := c.do(t, method, path, token, body)
	if resp.StatusCode != want {
		t.Fatalf("%s %s: got %d, want %d. body=%s", method, path, resp.StatusCode, want, string(raw))
	}
	if into != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, into); err != nil {
			t.Fatalf("%s %s: response is not JSON: %v. body=%s", method, path, err, string(raw))
		}
	}
}

func (c *apiClient) postExpect(t *testing.T, want int, path, token string, body, into any) {
	t.Helper()
	c.expect(t, http.MethodPost, path, token, body, into, want)
}

func (c *apiClient) getExpect(t *testing.T, want int, path, token string, into any) {
	t.Helper()
	c.expect(t, http.MethodGet, path, token, nil, into, want)
}

func (c *apiClient) patchExpect(t *testing.T, want int, path, token string, body, into any) {
	t.Helper()
	c.expect(t, http.MethodPatch, path, token, body, into, want)
}
