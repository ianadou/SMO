package ports

// IDGenerator produces unique string identifiers for new entities.
//
// Implementations are typically thin wrappers around UUID libraries
// (e.g., github.com/google/uuid) but the interface stays free of any
// concrete dependency so the domain remains pure.
//
// Use cases inject this port to keep ID generation deterministic in
// tests: a test passes a fake generator that returns a known sequence
// of IDs, while production code passes a real UUID-based generator.
type IDGenerator interface {
	// Generate returns a new unique identifier. The returned value is
	// expected to be globally unique with extremely high probability
	// (e.g., a UUID v4).
	Generate() string
}
