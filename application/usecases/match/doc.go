// Package match contains the use cases that operate on the Match
// aggregate: creating a match, retrieving it, listing matches by group,
// and orchestrating the state machine transitions (open, mark teams
// ready, start, complete, close).
//
// State transition use cases live in their own file (transitions.go)
// because they share a common pattern: load the match, invoke a domain
// state machine method, persist the new status.
//
// As with the group use cases, these are pure orchestrators: all
// business rules (status transitions, field validation) live in the
// Match entity; all persistence lives behind the MatchRepository port.
package match
