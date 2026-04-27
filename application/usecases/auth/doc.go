// Package auth holds the use cases that orchestrate the authentication
// flow for organizers: registration with an email + password, and login
// in exchange for a JWT.
//
// Players are NOT authenticated here: they are token-only participants
// (see the invitation aggregate), and their identity is checked through
// invitation tokens, not JWTs.
package auth
