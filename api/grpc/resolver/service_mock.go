package resolver

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/event"
)

type serviceMock struct {
	hasCapacity bool
}

func NewServiceMock() Service {
	return &serviceMock{}
}

func (sm *serviceMock) SubmitBatch(ctx context.Context, msgs []*event.Event) (count uint32, err error) {
	for _, msg := range msgs {
		if msg.ID() == "resolver_fail" {
			err = ErrInternal
			break
		}
		if msg.ID() == "resolver_queue_missing" {
			err = ErrQueueMissing
			break
		}
		if msg.ID() == "full" && !sm.hasCapacity {
			sm.hasCapacity = true
			break
		}
		count++
	}
	return
}
