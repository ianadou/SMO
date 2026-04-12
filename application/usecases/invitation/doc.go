// Package invitation contains the use cases that operate on the
// Invitation aggregate: creating a token-backed invitation, retrieving
// it, listing invitations per match, and accepting one by consuming
// its token.
//
// CreateInvitationUseCase is the only place where the plain token is
// visible: it is returned once to the caller and never persisted.
// Subsequent reads only ever see the hash.
package invitation
