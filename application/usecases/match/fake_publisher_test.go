package match

import (
	"context"

	"github.com/ianadou/smo/domain/events"
)

// fakePublisher records every event passed to Publish so tests can
// assert that the use case emitted what it should have. It implements
// ports.EventPublisher.
type fakePublisher struct {
	published []events.Event
}

func newFakePublisher() *fakePublisher {
	return &fakePublisher{}
}

func (p *fakePublisher) Publish(_ context.Context, event events.Event) {
	p.published = append(p.published, event)
}
