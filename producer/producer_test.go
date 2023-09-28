package producer

import (
	"context"
	"github.com/SlyMarbo/rss"
	"github.com/awakari/client-sdk-go/model"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
	"os"
	"producer-rss/config"
	"producer-rss/converter"
	"testing"
	"time"
)

func TestProducer_Produce(t *testing.T) {
	feed := &rss.Feed{
		Nickname:    "test-feed-name-0",
		Title:       "test-feed-title-0",
		Language:    "mi-NZ",
		Author:      "test-feed-author-0",
		Description: "test-feed-description-0",
		Link:        "https://test-feed-0.nz",
		UpdateURL:   "https://test-feed-0.nz",
		Image:       nil,
		Categories: []string{
			"test",
			"feed",
		},
		Items: []*rss.Item{
			{
				Title:   "item-0-title",
				Summary: "item-0-summary",
				Content: "item-0-content",
				Link:    "https://test-feed-0.nz/item0",
				Date:    time.Date(2023, 6, 9, 7, 31, 50, 0, time.UTC),
			},
			{
				Title:   "item-1-title",
				Summary: "item-1-summary",
				Content: "item-1-content",
				Link:    "https://test-feed-0.nz/item1",
				Date:    time.Date(2023, 6, 9, 7, 32, 50, 0, time.UTC),
			},
		},
	}
	os.Setenv("FEED_URL", "https://test-feed-0.nz")
	cfg, err := config.NewConfigFromEnv()
	require.Nil(t, err)
	conv := converter.NewConverter(cfg.Message)
	conv = converter.NewConverterLogging(conv, slog.Default())
	out := &testOutput{}
	timeMin := time.Date(2023, 6, 9, 7, 32, 0, 0, time.UTC)
	p := NewProducer(feed, timeMin, conv, out, 1*time.Second, 2)
	p = NewProducerLogging(p, slog.Default())
	var timeNext time.Time
	timeNext, err = p.Produce(context.TODO())
	assert.True(t, timeMin.Before(timeNext))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(out.Msgs))
	assert.Equal(t, "https://test-feed-0.nz", out.Msgs[0].Source)
	assert.Equal(t, "https://test-feed-0.nz/item1", out.Msgs[0].Attributes["subject"].GetCeString())
}

type testOutput struct {
	Msgs []*pb.CloudEvent
}

func (t *testOutput) Close() error {
	//TODO implement me
	panic("implement me")
}

func (t *testOutput) WriteBatch(items []*pb.CloudEvent) (ackCount uint32, err error) {
	t.Msgs = append(t.Msgs, items...)
	return uint32(len(items)), nil
}

var _ model.Writer[*pb.CloudEvent] = (*testOutput)(nil)
