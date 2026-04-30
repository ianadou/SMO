package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/auth"
	"github.com/ianadou/smo/application/usecases/group"
	"github.com/ianadou/smo/application/usecases/invitation"
	"github.com/ianadou/smo/application/usecases/match"
	"github.com/ianadou/smo/application/usecases/player"
	"github.com/ianadou/smo/application/usecases/vote"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/events"
	"github.com/ianadou/smo/domain/ranking"
	bcryptauth "github.com/ianadou/smo/infrastructure/auth/bcrypt"
	"github.com/ianadou/smo/infrastructure/http/handlers"
	"github.com/ianadou/smo/infrastructure/http/middlewares"
)

// --- minimal stubs only for the auth-required wiring test ----------

type authRequiredFakeGroupRepo struct{}

func (authRequiredFakeGroupRepo) Save(context.Context, *entities.Group) error {
	return domainerrors.ErrReferencedEntityNotFound
}

func (authRequiredFakeGroupRepo) FindByID(context.Context, entities.GroupID) (*entities.Group, error) {
	return nil, domainerrors.ErrGroupNotFound
}

func (authRequiredFakeGroupRepo) Delete(context.Context, entities.GroupID) error {
	return nil
}

func (authRequiredFakeGroupRepo) ListByOrganizer(context.Context, entities.OrganizerID) ([]*entities.Group, error) {
	return nil, nil
}

func (authRequiredFakeGroupRepo) Update(context.Context, *entities.Group) error {
	return nil
}

type authRequiredFakeInvitationRepo struct{}

func (authRequiredFakeInvitationRepo) Save(context.Context, *entities.Invitation) error {
	return domainerrors.ErrReferencedEntityNotFound
}

func (authRequiredFakeInvitationRepo) FindByID(context.Context, entities.InvitationID) (*entities.Invitation, error) {
	return nil, domainerrors.ErrInvitationNotFound
}

func (authRequiredFakeInvitationRepo) FindByTokenHash(context.Context, string) (*entities.Invitation, error) {
	return nil, domainerrors.ErrInvitationNotFound
}

func (authRequiredFakeInvitationRepo) MarkAsUsed(context.Context, *entities.Invitation) error {
	return domainerrors.ErrInvitationNotFound
}

func (authRequiredFakeInvitationRepo) ListByMatch(context.Context, entities.MatchID) ([]*entities.Invitation, error) {
	return nil, nil
}

func (authRequiredFakeInvitationRepo) Delete(context.Context, entities.InvitationID) error {
	return nil
}

type authRequiredFakeOrganizerRepo struct{}

func (authRequiredFakeOrganizerRepo) Save(context.Context, *entities.Organizer) error {
	return domainerrors.ErrEmailAlreadyExists
}

func (authRequiredFakeOrganizerRepo) FindByID(context.Context, entities.OrganizerID) (*entities.Organizer, error) {
	return nil, domainerrors.ErrOrganizerNotFound
}

func (authRequiredFakeOrganizerRepo) FindByEmail(context.Context, string) (*entities.Organizer, error) {
	return nil, domainerrors.ErrOrganizerNotFound
}

type authRequiredFakePlayerRepo struct{}

func (authRequiredFakePlayerRepo) Save(context.Context, *entities.Player) error {
	return domainerrors.ErrReferencedEntityNotFound
}

func (authRequiredFakePlayerRepo) FindByID(context.Context, entities.PlayerID) (*entities.Player, error) {
	return nil, domainerrors.ErrPlayerNotFound
}

func (authRequiredFakePlayerRepo) ListByGroup(context.Context, entities.GroupID) ([]*entities.Player, error) {
	return nil, nil
}

func (authRequiredFakePlayerRepo) UpdateRanking(context.Context, *entities.Player) error {
	return domainerrors.ErrPlayerNotFound
}

func (authRequiredFakePlayerRepo) Delete(context.Context, entities.PlayerID) error {
	return nil
}

type authRequiredFakeVoteRepo struct{}

func (authRequiredFakeVoteRepo) Save(context.Context, *entities.Vote) error {
	return domainerrors.ErrReferencedEntityNotFound
}

func (authRequiredFakeVoteRepo) FindByID(context.Context, entities.VoteID) (*entities.Vote, error) {
	return nil, domainerrors.ErrVoteNotFound
}

func (authRequiredFakeVoteRepo) ListByMatch(context.Context, entities.MatchID) ([]*entities.Vote, error) {
	return nil, nil
}

func (authRequiredFakeVoteRepo) ListByVoter(context.Context, entities.PlayerID) ([]*entities.Vote, error) {
	return nil, nil
}

func (authRequiredFakeVoteRepo) Delete(context.Context, entities.VoteID) error { return nil }

type authRequiredFakeMatchRepo struct{}

func (authRequiredFakeMatchRepo) Save(context.Context, *entities.Match) error {
	return domainerrors.ErrReferencedEntityNotFound
}

func (authRequiredFakeMatchRepo) FindByID(context.Context, entities.MatchID) (*entities.Match, error) {
	return nil, domainerrors.ErrMatchNotFound
}

func (authRequiredFakeMatchRepo) ListByGroup(context.Context, entities.GroupID) ([]*entities.Match, error) {
	return nil, nil
}

func (authRequiredFakeMatchRepo) UpdateStatus(context.Context, *entities.Match) error {
	return domainerrors.ErrMatchNotFound
}

func (authRequiredFakeMatchRepo) Finalize(context.Context, *entities.Match) error {
	return domainerrors.ErrMatchNotFound
}

func (authRequiredFakeMatchRepo) Delete(context.Context, entities.MatchID) error { return nil }

type noopTokenService struct{}

func (noopTokenService) GenerateToken() (string, error) { return "stub-token", nil }
func (noopTokenService) HashToken(string) string        { return "stub-hash" }

// noopPublisher discards every Publish call. Handler middleware tests
// only care about HTTP wiring, not domain event delivery.
type noopPublisher struct{}

func (noopPublisher) Publish(context.Context, events.Event) {}

// authRequiredStubSigner is a minimal JWTSigner stub: it accepts any
// token whose value is "valid-token" and rejects everything else.
// Used to test the JWTAuth wiring without depending on real JWT logic.
type authRequiredStubSigner struct{}

func (authRequiredStubSigner) Sign(_ entities.OrganizerID) (string, error) {
	return "valid-token", nil
}

func (authRequiredStubSigner) Verify(token string) (entities.OrganizerID, error) {
	if token == "valid-token" {
		return "org-1", nil
	}
	return "", domainerrors.ErrInvalidToken
}

// buildProtectedRouter wires the full router with public/protected
// groups and the JWTAuth middleware on the protected group, exactly as
// main.go does. Handlers receive minimal dependencies that panic on use:
// the test only checks middleware-level behavior (401 vs handler).
func buildProtectedRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := authRequiredFakeMatchRepo{}
	voteRepo := authRequiredFakeVoteRepo{}
	playerRepo := authRequiredFakePlayerRepo{}
	idGen := &fixedIDGen{id: "id-stub"}
	clock := &fixedClock{}

	calculator, err := ranking.NewCalculator(ranking.DefaultLearningRate())
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}
	hasher := bcryptauth.New(4)

	matchHandler := handlers.NewMatchHandler(
		match.NewCreateMatchUseCase(repo, idGen, clock),
		match.NewGetMatchUseCase(repo),
		match.NewListMatchesByGroupUseCase(repo),
		match.NewOpenMatchUseCase(repo),
		match.NewMarkTeamsReadyUseCase(repo, noopPublisher{}, clock),
		match.NewStartMatchUseCase(repo),
		match.NewCompleteMatchUseCase(repo),
		match.NewFinalizeMatchUseCase(repo, voteRepo, playerRepo, calculator),
	)

	groupRepo := &authRequiredFakeGroupRepo{}
	groupHandler := handlers.NewGroupHandler(
		group.NewCreateGroupUseCase(groupRepo, idGen, clock),
		group.NewGetGroupUseCase(groupRepo),
		group.NewListGroupsByOrganizerUseCase(groupRepo),
	)

	invitationHandler := handlers.NewInvitationHandler(
		invitation.NewCreateInvitationUseCase(&authRequiredFakeInvitationRepo{}, &noopTokenService{}, idGen, clock),
		invitation.NewGetInvitationUseCase(&authRequiredFakeInvitationRepo{}),
		invitation.NewListInvitationsByMatchUseCase(&authRequiredFakeInvitationRepo{}),
		invitation.NewAcceptInvitationUseCase(&authRequiredFakeInvitationRepo{}, &noopTokenService{}, clock),
	)

	playerHandler := handlers.NewPlayerHandler(
		player.NewCreatePlayerUseCase(playerRepo, idGen),
		player.NewGetPlayerUseCase(playerRepo),
		player.NewListPlayersByGroupUseCase(playerRepo),
		player.NewUpdatePlayerRankingUseCase(playerRepo),
	)

	voteHandler := handlers.NewVoteHandler(
		vote.NewCastVoteUseCase(voteRepo, repo, idGen, clock),
		vote.NewGetVoteUseCase(voteRepo),
		vote.NewListVotesByMatchUseCase(voteRepo),
	)

	authHandler := handlers.NewAuthHandler(
		auth.NewRegisterOrganizerUseCase(&authRequiredFakeOrganizerRepo{}, hasher, idGen, clock),
		auth.NewLoginOrganizerUseCase(&authRequiredFakeOrganizerRepo{}, hasher, authRequiredStubSigner{}),
	)

	router := gin.New()
	public := router.Group("/api/v1")
	protected := router.Group("/api/v1")
	protected.Use(middlewares.JWTAuth(authRequiredStubSigner{}))

	groupHandler.Register(public, protected)
	matchHandler.Register(public, protected)
	playerHandler.Register(public, protected)
	invitationHandler.Register(public, protected)
	voteHandler.Register(public, protected)
	authHandler.Register(public, protected)

	return router
}

func TestAuthRequired_ProtectedRoute_Returns401_WithoutToken(t *testing.T) {
	t.Parallel()
	router := buildProtectedRouter(t)

	cases := []struct {
		method, path string
	}{
		{http.MethodPost, "/api/v1/groups"},
		{http.MethodPost, "/api/v1/matches"},
		{http.MethodPost, "/api/v1/matches/m-1/open"},
		{http.MethodPost, "/api/v1/players"},
		{http.MethodPatch, "/api/v1/players/p-1/ranking"},
		{http.MethodPost, "/api/v1/invitations"},
		{http.MethodGet, "/api/v1/matches/m-1/invitations"},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d (body=%s)", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAuthRequired_PublicRoute_DoesNotReturn401_WithoutToken(t *testing.T) {
	t.Parallel()
	router := buildProtectedRouter(t)

	// Public routes must NOT return 401 without a token. They may
	// return any other status (404, 400, 200…) depending on the handler
	// logic — the only thing that matters here is that JWTAuth did not
	// reject them.
	cases := []struct {
		method, path string
	}{
		{http.MethodGet, "/api/v1/groups/g-1"},
		{http.MethodGet, "/api/v1/matches/m-1"},
		{http.MethodGet, "/api/v1/players/p-1"},
		{http.MethodGet, "/api/v1/groups/g-1/players"},
		{http.MethodGet, "/api/v1/groups/g-1/matches"},
		{http.MethodGet, "/api/v1/votes/v-1"},
		{http.MethodGet, "/api/v1/matches/m-1/votes"},
		{http.MethodGet, "/api/v1/invitations/i-1"},
		{http.MethodPost, "/api/v1/auth/register"},
		{http.MethodPost, "/api/v1/auth/login"},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code == http.StatusUnauthorized {
				t.Errorf("public route unexpectedly returned 401 (body=%s)", rec.Body.String())
			}
		})
	}
}
