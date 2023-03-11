package resolver

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/event"
)

type serviceMock struct {
}

func NewServiceMock() Service {
	return serviceMock{}
}

func (sm serviceMock) Submit(ctx context.Context, msg *event.Event) (err error) {
	switch msg.ID() {
	case "resolver_fail":
		return ErrInternal
	case "resolver_queue_full":
		return ErrQueueFull
	case "resolver_queue_missing":
		return ErrQueueMissing
	}
	return
}
