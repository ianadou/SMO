package group

// fakeIDGenerator is a deterministic implementation of the IDGenerator
// port that returns a fixed sequence of IDs. Tests use it to predict
// exactly which IDs will be assigned to entities created during the test.
type fakeIDGenerator struct {
	ids   []string
	index int
}

func newFakeIDGenerator(ids ...string) *fakeIDGenerator {
	return &fakeIDGenerator{ids: ids}
}

func (g *fakeIDGenerator) Generate() string {
	if g.index >= len(g.ids) {
		// Tests should always provide enough IDs; running out is a
		// test setup bug, not a runtime concern.
		panic("fakeIDGenerator: ran out of IDs")
	}
	id := g.ids[g.index]
	g.index++
	return id
}
