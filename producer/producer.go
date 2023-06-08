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
	feed            *rss.Feed
	timeMin         time.Time
	conv            converter.Converter
	output          model.Writer[*pb.CloudEvent]
	outputBackoff   time.Duration
	outputBatchSize uint32
}

func NewProducer(feed *rss.Feed, timeMin time.Time, conv converter.Converter, output model.Writer[*pb.CloudEvent], outputBackoff time.Duration, outputBatchSize uint32) Producer {
	return producer{
		feed:            feed,
		timeMin:         timeMin,
		conv:            conv,
		output:          output,
		outputBackoff:   outputBackoff,
		outputBatchSize: outputBatchSize,
	}
}

func (p producer) Produce(ctx context.Context) (timeMax time.Time, err error) {
	var msgBatch []*pb.CloudEvent
	var msg *pb.CloudEvent
	for _, item := range p.feed.Items {
		if item.Date.IsZero() || p.timeMin.Before(item.Date) {
			msg = p.conv.Convert(p.feed, item)
			msgBatch = append(msgBatch, msg)
			if uint32(len(msgBatch)) == p.outputBatchSize {
				// flush
				err = errors.Join(err, p.sendMessages(ctx, msgBatch))
				msgBatch = []*pb.CloudEvent{}
			}
		}
		if item.DateValid && timeMax.Before(item.Date) {
			timeMax = item.Date
		}
	}
	// send the remaining messages, if any
	if len(msgBatch) > 0 {
		err = errors.Join(err, p.sendMessages(ctx, msgBatch))
	}
	if timeMax.IsZero() {
		timeMax = time.Now().UTC()
	}
	return
}

func (p producer) sendMessages(ctx context.Context, msgs []*pb.CloudEvent) (err error) {
	msgCount := uint32(len(msgs))
	var ackCount uint32
	var n uint32
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
