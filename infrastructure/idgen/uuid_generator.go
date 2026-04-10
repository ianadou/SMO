package idgen

import "github.com/google/uuid"

// UUIDGenerator generates random UUID v4 strings. It implements the
// domain ports.IDGenerator interface.
//
// UUID v4 is suitable for entity IDs because it provides 122 bits of
// entropy, making collisions practically impossible without any need
// for coordination between processes.
type UUIDGenerator struct{}

// New returns a new UUIDGenerator. There is no configuration to provide;
// the constructor exists for symmetry with other adapters and to make
// the dependency explicit at the composition root.
func New() *UUIDGenerator {
	return &UUIDGenerator{}
}

// Generate returns a new random UUID v4 as a string in the canonical
// 8-4-4-4-12 hexadecimal format.
func (g *UUIDGenerator) Generate() string {
	return uuid.NewString()
}
