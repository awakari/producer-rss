package resolver

import (
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_Route(t *testing.T) {
	svc := NewService(NewClientMock())
	cases := map[string]error{
		"fail":                                 ErrInternal,
		"full":                                 ErrQueueFull,
		"missing":                              ErrQueueMissing,
		"3426d090-1b8a-4a09-ac9c-41f2de24d5ac": nil,
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			msg := event.New()
			msg.SetID(k)
			msg.SetData("application/octet-stream", []byte{42})
			err := svc.Submit(nil, &msg)
			assert.ErrorIs(t, err, c)
		})
	}
}
