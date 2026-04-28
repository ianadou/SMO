package errors

import (
	"errors"
	"net/http"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

// MapError translates a domain error into an HTTP status code and a
// safe public message. Unknown errors are reported as 500 with a
// generic message; the original error should be logged separately by
// the caller for diagnosis.
//
// The mapping is split into helper functions per error category
// (validation, business rules, not-found) to keep cyclomatic
// complexity within the project linter limits.
func MapError(err error) (int, string) {
	if status, message, matched := mapNotFoundError(err); matched {
		return status, message
	}
	if status, message, matched := mapValidationError(err); matched {
		return status, message
	}
	if status, message, matched := mapBusinessRuleError(err); matched {
		return status, message
	}
	if status, message, matched := mapAuthError(err); matched {
		return status, message
	}
	return http.StatusInternalServerError, "internal server error"
}

// mapNotFoundError handles errors that translate to HTTP 404.
func mapNotFoundError(err error) (int, string, bool) {
	if errors.Is(err, domainerrors.ErrGroupNotFound) {
		return http.StatusNotFound, "group not found", true
	}
	if errors.Is(err, domainerrors.ErrMatchNotFound) {
		return http.StatusNotFound, "match not found", true
	}
	if errors.Is(err, domainerrors.ErrPlayerNotFound) {
		return http.StatusNotFound, "player not found", true
	}
	if errors.Is(err, domainerrors.ErrInvitationNotFound) {
		return http.StatusNotFound, "invitation not found", true
	}
	if errors.Is(err, domainerrors.ErrVoteNotFound) {
		return http.StatusNotFound, "vote not found", true
	}
	if errors.Is(err, domainerrors.ErrOrganizerNotFound) {
		return http.StatusNotFound, "organizer not found", true
	}
	return 0, "", false
}

// mapValidationError handles errors that translate to HTTP 400 because
// the input did not pass format or value validation, or because it
// referenced an entity that does not exist.
func mapValidationError(err error) (int, string, bool) {
	switch {
	case errors.Is(err, domainerrors.ErrInvalidID):
		return http.StatusBadRequest, "invalid id", true
	case errors.Is(err, domainerrors.ErrInvalidName):
		return http.StatusBadRequest, "invalid name", true
	case errors.Is(err, domainerrors.ErrInvalidScore):
		return http.StatusBadRequest, "invalid score", true
	case errors.Is(err, domainerrors.ErrInvalidDate):
		return http.StatusBadRequest, "invalid date", true
	case errors.Is(err, domainerrors.ErrInvalidStatus):
		return http.StatusBadRequest, "invalid status", true
	case errors.Is(err, domainerrors.ErrInvalidParameter):
		return http.StatusBadRequest, "invalid parameter", true
	case errors.Is(err, domainerrors.ErrReferencedEntityNotFound):
		return http.StatusBadRequest, "referenced entity does not exist", true
	case errors.Is(err, domainerrors.ErrInvalidEmail):
		return http.StatusBadRequest, "invalid email", true
	case errors.Is(err, domainerrors.ErrInvalidPassword):
		return http.StatusBadRequest, "invalid password", true
	case errors.Is(err, domainerrors.ErrInvalidWebhookURL):
		return http.StatusBadRequest, "invalid webhook url", true
	}
	return 0, "", false
}

// mapBusinessRuleError handles errors that come from a business rule
// violation: the request was syntactically valid but the operation is
// not allowed in the current state of the resource.
func mapBusinessRuleError(err error) (int, string, bool) {
	switch {
	case errors.Is(err, domainerrors.ErrInvalidTransition):
		return http.StatusConflict, "operation not allowed in current state", true
	case errors.Is(err, domainerrors.ErrInvalidAssignment):
		return http.StatusBadRequest, "invalid team assignment", true
	case errors.Is(err, domainerrors.ErrSelfVote):
		return http.StatusBadRequest, "cannot vote for yourself", true
	case errors.Is(err, domainerrors.ErrMatchFull):
		return http.StatusConflict, "match is full", true
	case errors.Is(err, domainerrors.ErrTeamFull):
		return http.StatusConflict, "team is full", true
	case errors.Is(err, domainerrors.ErrPlayerNotInMatch):
		return http.StatusBadRequest, "player is not in this match", true
	case errors.Is(err, domainerrors.ErrInvitationExpired):
		return http.StatusGone, "invitation expired", true
	case errors.Is(err, domainerrors.ErrInvitationAlreadyUsed):
		return http.StatusConflict, "invitation already used", true
	case errors.Is(err, domainerrors.ErrAlreadyVoted):
		return http.StatusConflict, "already voted for this player in this match", true
	case errors.Is(err, domainerrors.ErrMatchNotCompleted):
		return http.StatusConflict, "match is not completed", true
	case errors.Is(err, domainerrors.ErrEmailAlreadyExists):
		return http.StatusConflict, "email already exists", true
	}
	return 0, "", false
}

// mapAuthError handles errors specific to the authentication flow.
// Both ErrInvalidCredentials and ErrInvalidToken map to 401 with a
// short message so the response shape never reveals whether the email
// exists or whether the token was malformed vs expired.
func mapAuthError(err error) (int, string, bool) {
	switch {
	case errors.Is(err, domainerrors.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid credentials", true
	case errors.Is(err, domainerrors.ErrInvalidToken):
		return http.StatusUnauthorized, "invalid token", true
	}
	return 0, "", false
}

// ErrorResponse is the JSON shape returned by handlers when an error
// occurs. It is intentionally minimal: just a human-readable message,
// no stack trace, no internal details.
type ErrorResponse struct {
	Error string `json:"error"`
}
