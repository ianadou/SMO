// Package entities defines the core domain entities of the SMO project.
//
// Each entity is an immutable struct exposed through getter methods only.
// All construction goes through a NewXxx constructor that validates inputs
// and rejects invalid states with sentinel errors from the domain/errors
// package.
//
// Entities in this package have no dependencies on persistence, transport,
// or any other infrastructure concern. They are pure Go structures that
// can be tested in isolation.
package entities
