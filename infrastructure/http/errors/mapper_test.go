package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestMapError_NotFoundErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		err         error
		wantMessage string
	}{
		{name: "group", err: domainerrors.ErrGroupNotFound, wantMessage: "group not found"},
		{name: "match", err: domainerrors.ErrMatchNotFound, wantMessage: "match not found"},
		{name: "player", err: domainerrors.ErrPlayerNotFound, wantMessage: "player not found"},
		{name: "invitation", err: domainerrors.ErrInvitationNotFound, wantMessage: "invitation not found"},
		{name: "vote", err: domainerrors.ErrVoteNotFound, wantMessage: "vote not found"},
		{name: "organizer", err: domainerrors.ErrOrganizerNotFound, wantMessage: "organizer not found"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			status, message := MapError(testCase.err)
			if status != http.StatusNotFound {
				t.Errorf("expected status 404, got %d", status)
			}
			if message != testCase.wantMessage {
				t.Errorf("expected message %q, got %q", testCase.wantMessage, message)
			}
		})
	}
}

func TestMapError_ValidationErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{name: "invalid id", err: domainerrors.ErrInvalidID, wantStatus: http.StatusBadRequest, wantMessage: "invalid id"},
		{name: "invalid name", err: domainerrors.ErrInvalidName, wantStatus: http.StatusBadRequest, wantMessage: "invalid name"},
		{name: "invalid score", err: domainerrors.ErrInvalidScore, wantStatus: http.StatusBadRequest, wantMessage: "invalid score"},
		{name: "invalid date", err: domainerrors.ErrInvalidDate, wantStatus: http.StatusBadRequest, wantMessage: "invalid date"},
		{name: "invalid status", err: domainerrors.ErrInvalidStatus, wantStatus: http.StatusBadRequest, wantMessage: "invalid status"},
		{name: "invalid parameter", err: domainerrors.ErrInvalidParameter, wantStatus: http.StatusBadRequest, wantMessage: "invalid parameter"},
		{name: "referenced entity not found", err: domainerrors.ErrReferencedEntityNotFound, wantStatus: http.StatusBadRequest, wantMessage: "referenced entity does not exist"},
		{name: "invalid email", err: domainerrors.ErrInvalidEmail, wantStatus: http.StatusBadRequest, wantMessage: "invalid email"},
		{name: "invalid password", err: domainerrors.ErrInvalidPassword, wantStatus: http.StatusBadRequest, wantMessage: "invalid password"},
		{name: "invalid webhook url", err: domainerrors.ErrInvalidWebhookURL, wantStatus: http.StatusBadRequest, wantMessage: "invalid webhook url"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			status, message := MapError(testCase.err)

			if status != testCase.wantStatus {
				t.Errorf("expected status %d, got %d", testCase.wantStatus, status)
			}
			if message != testCase.wantMessage {
				t.Errorf("expected message %q, got %q", testCase.wantMessage, message)
			}
		})
	}
}

func TestMapError_BusinessRuleErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{name: "invalid transition", err: domainerrors.ErrInvalidTransition, wantStatus: http.StatusConflict, wantMessage: "operation not allowed in current state"},
		{name: "invalid assignment", err: domainerrors.ErrInvalidAssignment, wantStatus: http.StatusBadRequest, wantMessage: "invalid team assignment"},
		{name: "self vote", err: domainerrors.ErrSelfVote, wantStatus: http.StatusBadRequest, wantMessage: "cannot vote for yourself"},
		{name: "match full", err: domainerrors.ErrMatchFull, wantStatus: http.StatusConflict, wantMessage: "match is full"},
		{name: "team full", err: domainerrors.ErrTeamFull, wantStatus: http.StatusConflict, wantMessage: "team is full"},
		{name: "player not in match", err: domainerrors.ErrPlayerNotInMatch, wantStatus: http.StatusBadRequest, wantMessage: "player is not in this match"},
		{name: "invitation expired", err: domainerrors.ErrInvitationExpired, wantStatus: http.StatusGone, wantMessage: "invitation expired"},
		{name: "invitation already used", err: domainerrors.ErrInvitationAlreadyUsed, wantStatus: http.StatusConflict, wantMessage: "invitation already used"},
		{name: "already voted", err: domainerrors.ErrAlreadyVoted, wantStatus: http.StatusConflict, wantMessage: "already voted for this player in this match"},
		{name: "match not completed", err: domainerrors.ErrMatchNotCompleted, wantStatus: http.StatusConflict, wantMessage: "match is not completed"},
		{name: "email already exists", err: domainerrors.ErrEmailAlreadyExists, wantStatus: http.StatusConflict, wantMessage: "email already exists"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			status, message := MapError(testCase.err)

			if status != testCase.wantStatus {
				t.Errorf("expected status %d, got %d", testCase.wantStatus, status)
			}
			if message != testCase.wantMessage {
				t.Errorf("expected message %q, got %q", testCase.wantMessage, message)
			}
		})
	}
}

func TestMapError_AuthErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		err         error
		wantMessage string
	}{
		{name: "invalid credentials", err: domainerrors.ErrInvalidCredentials, wantMessage: "invalid credentials"},
		{name: "invalid token", err: domainerrors.ErrInvalidToken, wantMessage: "invalid token"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			status, message := MapError(testCase.err)
			if status != http.StatusUnauthorized {
				t.Errorf("expected status 401, got %d", status)
			}
			if message != testCase.wantMessage {
				t.Errorf("expected message %q, got %q", testCase.wantMessage, message)
			}
		})
	}
}

func TestMapError_UnknownErrorReturns500(t *testing.T) {
	t.Parallel()

	unknownErr := errors.New("something completely unexpected")

	status, message := MapError(unknownErr)

	if status != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", status)
	}
	if message != "internal server error" {
		t.Errorf("expected message 'internal server error', got %q", message)
	}
}

func TestMapError_HandlesWrappedErrors(t *testing.T) {
	t.Parallel()

	// MapError must work with errors wrapped via fmt.Errorf("...: %w", ...)
	// because that is the standard pattern in our use cases and repositories.
	wrapped := fmt.Errorf("use case: do something: %w", domainerrors.ErrGroupNotFound)

	status, message := MapError(wrapped)

	if status != http.StatusNotFound {
		t.Errorf("expected status 404 for wrapped not-found error, got %d", status)
	}
	if message != "group not found" {
		t.Errorf("expected message 'group not found', got %q", message)
	}
}

func TestMapError_HandlesWrappedReferencedEntityNotFound(t *testing.T) {
	t.Parallel()

	// This is the exact wrapping pattern used by the Postgres repository
	// when detecting a foreign key violation during a Save call. The
	// mapper must unwrap correctly through both the repository layer and
	// the use case layer to return a 400.
	wrapped := fmt.Errorf(
		"create group use case: save group %q: %w",
		"group-1",
		fmt.Errorf(
			"postgres group repository: save group %q: %w",
			"group-1", domainerrors.ErrReferencedEntityNotFound,
		),
	)

	status, message := MapError(wrapped)

	if status != http.StatusBadRequest {
		t.Errorf("expected status 400 for wrapped referenced-entity-not-found, got %d", status)
	}
	if message != "referenced entity does not exist" {
		t.Errorf("expected message 'referenced entity does not exist', got %q", message)
	}
}
