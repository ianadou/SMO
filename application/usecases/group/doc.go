// Package group contains the use cases that operate on the Group
// aggregate: creating a group, retrieving it, listing groups by
// organizer, updating, and deleting.
//
// Each use case is a thin orchestrator that:
//  1. validates input format,
//  2. delegates ID generation and time retrieval to injected ports,
//  3. builds the domain entity through its constructor (which enforces
//     business invariants),
//  4. persists the entity through the GroupRepository port,
//  5. returns the entity to the caller.
//
// Use cases never contain business logic; that lives in the entities.
// They never know about HTTP, JSON, SQL, or any other infrastructure
// concern; they depend only on domain entities and ports.
package group
