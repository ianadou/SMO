// Package sharelink contains the use cases around the match share link:
// the single URL an organizer drops in the group chat so invitees can
// claim their personal invitation or add themselves to the roster.
//
// The organizer-facing use cases (generate, revoke) enforce ownership of
// the match's group. The public use cases (context, claim, join) resolve
// the bearer's plain token by hash and never reveal whether a dead link
// was revoked, expired, or never existed.
//
// GenerateMatchShareLinkUseCase, ClaimInvitationUseCase and
// JoinMatchUseCase are the only places where plain tokens are visible:
// they are returned once to the caller and never persisted.
package sharelink
