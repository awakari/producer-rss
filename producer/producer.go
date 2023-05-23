package producer

import (
	"context"
	"errors"
	"github.com/SlyMarbo/rss"
	"github.com/awakari/client-sdk-go/model"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"producer-rss/converter"
	"time"
)

type Producer interface {
	Produce(ctx context.Context) (nextTime time.Time, err error)
}

type producer struct {
	feed          *rss.Feed
	timeMin       time.Time
	conv          converter.Converter
	output        model.WriteStream[*pb.CloudEvent]
	outputBackoff time.Duration
}

func NewProducer(feed *rss.Feed, timeMin time.Time, conv converter.Converter, output model.WriteStream[*pb.CloudEvent], outputBackoff time.Duration) Producer {
	return producer{
		feed:          feed,
		timeMin:       timeMin,
		conv:          conv,
		output:        output,
		outputBackoff: outputBackoff,
	}
}

func (p producer) Produce(ctx context.Context) (timeMax time.Time, err error) {
	var msgs []*pb.CloudEvent
	msgs, timeMax, err = p.getNewMessages()
	if err == nil && len(msgs) > 0 {
		err = p.sendMessages(ctx, msgs)
	}
	return
}

func (p producer) getNewMessages() (msgs []*pb.CloudEvent, nextTime time.Time, err error) {
	var msg *pb.CloudEvent
	var itemErr error
	for _, item := range p.feed.Items {
		if item.Date.IsZero() || p.timeMin.Before(item.Date) {
			msg, itemErr = p.conv.Convert(p.feed, item)
			msgs = append(msgs, msg)
			err = errors.Join(err, itemErr)
		}
		if nextTime.Before(item.Date) {
			nextTime = item.Date
		}
	}
	return
}

func (p producer) sendMessages(ctx context.Context, msgs []*pb.CloudEvent) (err error) {
	msgCount := uint32(len(msgs))
	var ackCount, n uint32
	for ackCount < msgCount {
		n, err = p.output.WriteBatch(msgs[ackCount:])
		ackCount += n
		if err != nil {
			break
		}
		if n == 0 {
			time.Sleep(p.outputBackoff)
		}
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break
		default:
			continue
		}
	}
	return
}
