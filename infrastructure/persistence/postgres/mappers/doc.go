// Package mappers translates between sqlc-generated database structs
// and domain entities.
//
// The persistence layer uses two different representations of the same
// data:
//
//   - sqlc-generated structs (e.g., generated.Groups) which use
//     pgtype.Timestamptz, plain strings for IDs, and follow database
//     conventions.
//
//   - domain entities (e.g., entities.Group) which use time.Time, custom
//     ID types, and are immutable from outside their package.
//
// Mappers in this package convert between these two representations in
// both directions. They are pure functions with no IO and are easy to
// unit-test.
package mappers
