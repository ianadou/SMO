package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/vote"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/http/handlers"
)

// --- fakes ---------------------------------------------------------------

type fakeVoteRepo struct {
	mu    sync.Mutex
	votes map[entities.VoteID]*entities.Vote
}

func newFakeVoteRepo() *fakeVoteRepo {
	return &fakeVoteRepo{votes: make(map[entities.VoteID]*entities.Vote)}
}

func (r *fakeVoteRepo) Save(_ context.Context, v *entities.Vote) error {
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

func (r *fakeVoteRepo) FindByID(_ context.Context, id entities.VoteID) (*entities.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.votes[id]
	if !ok {
		return nil, domainerrors.ErrVoteNotFound
	}
	return v, nil
}

func (r *fakeVoteRepo) ListByMatch(_ context.Context, matchID entities.MatchID) ([]*entities.Vote, error) {
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

func (r *fakeVoteRepo) ListByVoter(_ context.Context, voterID entities.PlayerID) ([]*entities.Vote, error) {
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

func (r *fakeVoteRepo) Delete(_ context.Context, id entities.VoteID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.votes, id)
	return nil
}

// Minimal MatchRepository for the vote routes: lookup, roster read
// model and the matches-together aggregate.
type voteTestMatchRepo struct {
	match *entities.Match
}

func (r *voteTestMatchRepo) FindByID(_ context.Context, id entities.MatchID) (*entities.Match, error) {
	if r.match == nil || r.match.ID() != id {
		return nil, domainerrors.ErrMatchNotFound
	}
	return r.match, nil
}

func (r *voteTestMatchRepo) ListTeamMembersWithPlayers(context.Context, entities.MatchID) ([]entities.MatchTeamMember, error) {
	return []entities.MatchTeamMember{
		{PlayerID: "p-1", PlayerName: "Alice Martin", Team: "A", Slot: 0},
		{PlayerID: "p-2", PlayerName: "Bob Durand", Team: "A", Slot: 1},
		{PlayerID: "p-3", PlayerName: "Carol Petit", Team: "B", Slot: 0},
		{PlayerID: "p-4", PlayerName: "Dan Leroy", Team: "B", Slot: 1},
	}, nil
}

func (r *voteTestMatchRepo) CountClosedMatchesTogether(context.Context, entities.GroupID, entities.PlayerID, []entities.PlayerID) (map[entities.PlayerID]int, error) {
	return map[entities.PlayerID]int{"p-2": 7}, nil
}

func (r *voteTestMatchRepo) FindLatestDecidedByGroup(context.Context, entities.GroupID, entities.MatchID) (*entities.Match, error) {
	return nil, domainerrors.ErrMatchNotFound
}

func (r *voteTestMatchRepo) Save(context.Context, *entities.Match) error { panic("unused") }
func (r *voteTestMatchRepo) ListByGroup(context.Context, entities.GroupID) ([]*entities.Match, error) {
	panic("unused")
}
func (r *voteTestMatchRepo) UpdateStatus(context.Context, *entities.Match) error { panic("unused") }
func (r *voteTestMatchRepo) Finalize(context.Context, *entities.Match) error     { panic("unused") }
func (r *voteTestMatchRepo) ReplaceTeams(context.Context, *entities.Match) error { panic("unused") }
func (r *voteTestMatchRepo) Delete(context.Context, entities.MatchID) error      { panic("unused") }

// Minimal InvitationRepository keyed by token hash.
type voteTestInvitationRepo struct {
	invitations map[string]*entities.Invitation
}

func (r *voteTestInvitationRepo) FindByTokenHash(_ context.Context, hash string) (*entities.Invitation, error) {
	inv, ok := r.invitations[hash]
	if !ok {
		return nil, domainerrors.ErrInvitationNotFound
	}
	return inv, nil
}

func (r *voteTestInvitationRepo) Save(context.Context, *entities.Invitation) error { panic("unused") }
func (r *voteTestInvitationRepo) FindByID(context.Context, entities.InvitationID) (*entities.Invitation, error) {
	panic("unused")
}

func (r *voteTestInvitationRepo) ListByMatch(context.Context, entities.MatchID) ([]*entities.Invitation, error) {
	panic("unused")
}

func (r *voteTestInvitationRepo) CountConfirmedByMatch(context.Context, entities.MatchID) (int, error) {
	panic("unused")
}

func (r *voteTestInvitationRepo) ListConfirmedParticipants(context.Context, entities.MatchID) ([]entities.MatchParticipant, error) {
	panic("unused")
}

func (r *voteTestInvitationRepo) RespondWithCapacityGuard(context.Context, *entities.Invitation, int) error {
	panic("unused")
}

func (r *voteTestInvitationRepo) Claim(context.Context, *entities.Invitation) error {
	panic("unused")
}

func (r *voteTestInvitationRepo) Delete(context.Context, entities.InvitationID) error {
	panic("unused")
}

// Minimal GroupRepository returning a fixed group.
type voteTestGroupRepo struct{ group *entities.Group }

func (r *voteTestGroupRepo) FindByID(context.Context, entities.GroupID) (*entities.Group, error) {
	return r.group, nil
}

func (r *voteTestGroupRepo) Save(context.Context, *entities.Group) error   { panic("unused") }
func (r *voteTestGroupRepo) Update(context.Context, *entities.Group) error { panic("unused") }
func (r *voteTestGroupRepo) ListByOrganizer(context.Context, entities.OrganizerID) ([]*entities.Group, error) {
	panic("unused")
}
func (r *voteTestGroupRepo) Delete(context.Context, entities.GroupID) error { panic("unused") }

type voteTokenService struct{}

func (voteTokenService) GenerateToken() (string, error) { panic("unused") }
func (voteTokenService) HashToken(plain string) string  { return "hashed:" + plain }

type voteIDGen struct{ next int }

func (g *voteIDGen) Generate() string {
	g.next++
	return "v-gen-" + string(rune('0'+g.next))
}

type voteClock struct{ now time.Time }

func (c *voteClock) Now() time.Time { return c.now }

// --- helpers -------------------------------------------------------------

func buildVoteTestRouter(t *testing.T, matchStatus entities.MatchStatus) (*gin.Engine, *fakeVoteRepo) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	scoreA, scoreB := 3, 2
	snapshot := entities.MatchSnapshot{
		ID:          "test-match",
		GroupID:     "g-1",
		Title:       "Test",
		Venue:       "V",
		ScheduledAt: time.Date(2026, 6, 4, 19, 0, 0, 0, time.UTC),
		Status:      matchStatus,
		TeamA:       []entities.PlayerID{"p-1", "p-2"},
		TeamB:       []entities.PlayerID{"p-3", "p-4"},
		CreatedAt:   time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
	}
	if matchStatus == entities.MatchStatusCompleted || matchStatus == entities.MatchStatusClosed {
		snapshot.ScoreA = &scoreA
		snapshot.ScoreB = &scoreB
	}
	match, err := entities.RehydrateMatch(snapshot)
	if err != nil {
		t.Fatalf("seed match: %v", err)
	}

	respondedAt := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	confirmed, err := entities.NewInvitation("inv-1", "test-match", "p-1",
		voteTokenService{}.HashToken("tok-p-1"),
		time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC),
		entities.InvitationResponseYes, &respondedAt, nil,
		time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("seed confirmed invitation: %v", err)
	}
	declined, err := entities.NewInvitation("inv-2", "test-match", "p-2",
		voteTokenService{}.HashToken("tok-p-2"),
		time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC),
		entities.InvitationResponseNo, &respondedAt, nil,
		time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("seed declined invitation: %v", err)
	}
	invRepo := &voteTestInvitationRepo{invitations: map[string]*entities.Invitation{
		confirmed.TokenHash(): confirmed,
		declined.TokenHash():  declined,
	}}

	group, err := entities.NewGroup("g-1", "Foot du jeudi", "org-1", "",
		time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("seed group: %v", err)
	}

	voteRepo := newFakeVoteRepo()
	matchRepo := &voteTestMatchRepo{match: match}

	handler := handlers.NewVoteHandler(
		vote.NewCastVoteUseCase(voteRepo, matchRepo, invRepo, voteTokenService{},
			&voteIDGen{}, &voteClock{now: time.Date(2026, 6, 4, 21, 30, 0, 0, time.UTC)}),
		vote.NewGetVoteContextUseCase(invRepo, matchRepo, &voteTestGroupRepo{group: group},
			voteRepo, voteTokenService{}),
		vote.NewGetVoteUseCase(voteRepo),
		vote.NewListVotesByMatchUseCase(voteRepo),
	)
	router := gin.New()
	api := router.Group("/api/v1")
	handler.Register(api, api)
	return router, voteRepo
}

func postVoteJSON(t *testing.T, router *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// --- cast tests -----------------------------------------------------------

func TestVoteHandler_Cast_Returns201_WithVoterDerivedFromToken(t *testing.T) {
	router, _ := buildVoteTestRouter(t, entities.MatchStatusCompleted)

	rec := postVoteJSON(t, router, "/api/v1/votes",
		`{"token":"tok-p-1","voted_id":"p-2","score":4}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["voter_id"] != "p-1" {
		t.Errorf("expected voter_id derived from token to be p-1, got %v", resp["voter_id"])
	}
	if resp["score"].(float64) != 4 {
		t.Errorf("expected score 4, got %v", resp["score"])
	}
}

func TestVoteHandler_Cast_Returns404_WhenTokenUnknown(t *testing.T) {
	router, _ := buildVoteTestRouter(t, entities.MatchStatusCompleted)

	rec := postVoteJSON(t, router, "/api/v1/votes",
		`{"token":"tok-forged","voted_id":"p-2","score":4}`)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestVoteHandler_Cast_Returns403_WhenBearerDeclined(t *testing.T) {
	router, _ := buildVoteTestRouter(t, entities.MatchStatusCompleted)

	rec := postVoteJSON(t, router, "/api/v1/votes",
		`{"token":"tok-p-2","voted_id":"p-1","score":4}`)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestVoteHandler_Cast_Returns409_WhenMatchNotCompleted(t *testing.T) {
	router, _ := buildVoteTestRouter(t, entities.MatchStatusInProgress)

	rec := postVoteJSON(t, router, "/api/v1/votes",
		`{"token":"tok-p-1","voted_id":"p-2","score":4}`)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestVoteHandler_Cast_Returns400_OnSelfVote(t *testing.T) {
	router, _ := buildVoteTestRouter(t, entities.MatchStatusCompleted)

	rec := postVoteJSON(t, router, "/api/v1/votes",
		`{"token":"tok-p-1","voted_id":"p-1","score":4}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestVoteHandler_Cast_Returns400_WhenTargetIsOpponent(t *testing.T) {
	router, _ := buildVoteTestRouter(t, entities.MatchStatusCompleted)

	rec := postVoteJSON(t, router, "/api/v1/votes",
		`{"token":"tok-p-1","voted_id":"p-3","score":4}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

// --- context tests ----------------------------------------------------------

func TestVoteHandler_Context_Returns200_WithRatingView_WhenCompleted(t *testing.T) {
	router, _ := buildVoteTestRouter(t, entities.MatchStatusCompleted)

	rec := postVoteJSON(t, router, "/api/v1/votes/context", `{"token":"tok-p-1"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		GroupName string `json:"group_name"`
		Status    string `json:"status"`
		Winner    string `json:"winner"`
		Voter     struct {
			PlayerID string `json:"player_id"`
			Initials string `json:"initials"`
			Team     string `json:"team"`
		} `json:"voter"`
		Teammates []struct {
			PlayerID        string `json:"player_id"`
			Name            string `json:"name"`
			Initials        string `json:"initials"`
			MatchesTogether int    `json:"matches_together"`
			YourScore       *int   `json:"your_score"`
		} `json:"teammates"`
		VotersTotal int             `json:"voters_total"`
		Results     json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.GroupName != "Foot du jeudi" || resp.Status != "completed" || resp.Winner != "A" {
		t.Errorf("unexpected match block: %+v", resp)
	}
	if resp.Voter.PlayerID != "p-1" || resp.Voter.Initials != "AM" || resp.Voter.Team != "A" {
		t.Errorf("unexpected voter block: %+v", resp.Voter)
	}
	if len(resp.Teammates) != 1 || resp.Teammates[0].Name != "Bob Durand" ||
		resp.Teammates[0].Initials != "BD" || resp.Teammates[0].MatchesTogether != 7 {
		t.Errorf("unexpected teammates: %+v", resp.Teammates)
	}
	if resp.Teammates[0].YourScore != nil {
		t.Errorf("expected null your_score before voting, got %v", *resp.Teammates[0].YourScore)
	}
	if resp.VotersTotal != 4 {
		t.Errorf("expected voters_total 4, got %d", resp.VotersTotal)
	}
	if string(resp.Results) != "null" {
		t.Errorf("expected null results while completed, got %s", resp.Results)
	}
}

func TestVoteHandler_Context_Returns200_WithResults_WhenClosed(t *testing.T) {
	router, voteRepo := buildVoteTestRouter(t, entities.MatchStatusClosed)
	seeded, err := entities.NewVote("v-1", "test-match", "p-1", "p-2", 4,
		time.Date(2026, 6, 4, 22, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("seed vote: %v", err)
	}
	if saveErr := voteRepo.Save(context.Background(), seeded); saveErr != nil {
		t.Fatalf("save vote: %v", saveErr)
	}

	rec := postVoteJSON(t, router, "/api/v1/votes/context", `{"token":"tok-p-1"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		Results *struct {
			Teammates []struct {
				PlayerID   string   `json:"player_id"`
				Average    float64  `json:"average"`
				VotesCount int      `json:"votes_count"`
				Delta      *float64 `json:"delta"`
			} `json:"teammates"`
			Self struct {
				Average    *float64 `json:"average"`
				VotesCount int      `json:"votes_count"`
			} `json:"self"`
		} `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Results == nil {
		t.Fatal("expected results on a closed match")
	}
	teammate := resp.Results.Teammates[0]
	if teammate.PlayerID != "p-2" || teammate.Average != 4 || teammate.VotesCount != 1 {
		t.Errorf("unexpected teammate result: %+v", teammate)
	}
	if teammate.Delta != nil {
		t.Errorf("expected null delta without previous match, got %v", *teammate.Delta)
	}
	if resp.Results.Self.Average != nil {
		t.Errorf("expected null self average, got %v", *resp.Results.Self.Average)
	}
}

func TestVoteHandler_Context_Returns404_WhenTokenUnknown(t *testing.T) {
	router, _ := buildVoteTestRouter(t, entities.MatchStatusCompleted)

	rec := postVoteJSON(t, router, "/api/v1/votes/context", `{"token":"tok-forged"}`)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestVoteHandler_Context_Returns400_WhenTokenMissing(t *testing.T) {
	router, _ := buildVoteTestRouter(t, entities.MatchStatusCompleted)

	rec := postVoteJSON(t, router, "/api/v1/votes/context", `{}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}
