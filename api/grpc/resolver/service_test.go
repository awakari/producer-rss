package resolver

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_SubmitBatch(t *testing.T) {
	svc := NewService(NewClientMock())
	cases := map[string]struct {
		msgIds []string
		count  uint32
		err    error
	}{
		"ok": {
			msgIds: []string{
				"msg0",
				"msg1",
				"msg2",
			},
			count: 3,
		},
		"fail on 2nd": {
			msgIds: []string{
				"msg0",
				"fail",
				"msg2",
			},
			count: 1,
			err:   ErrInternal,
		},
		"not enough space in the queue": {
			msgIds: []string{
				"msg0",
				"msg1",
				"full",
			},
			count: 2,
		},
		"queue lost": {
			msgIds: []string{
				"missing",
				"msg1",
				"msg2",
			},
			count: 0,
			err:   ErrQueueMissing,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var msgs []*event.Event
			for _, msgId := range c.msgIds {
				msg := event.New()
				msg.SetID(msgId)
				msgs = append(msgs, &msg)
			}
			count, err := svc.SubmitBatch(context.TODO(), msgs)
			assert.Equal(t, c.count, count)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
