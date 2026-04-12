package vote

type fakeIDGenerator struct {
	ids   []string
	index int
}

func newFakeIDGenerator(ids ...string) *fakeIDGenerator {
	return &fakeIDGenerator{ids: ids}
}

func (g *fakeIDGenerator) Generate() string {
	if g.index >= len(g.ids) {
		panic("out of IDs")
	}
	id := g.ids[g.index]
	g.index++
	return id
}
