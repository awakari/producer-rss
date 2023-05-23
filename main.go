package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/SlyMarbo/rss"
	"github.com/awakari/client-sdk-go/api"
	"github.com/awakari/client-sdk-go/model"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"os"
	"producer-rss/config"
	"producer-rss/feeds"
	"strings"
	"time"
)

func main() {
	//
	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("failed to load the config from env: %s", err))
	}
	ctx := context.TODO()
	//
	opts := slog.HandlerOptions{
		Level: slog.Level(cfg.Log.Level),
	}
	log := slog.New(opts.NewTextHandler(os.Stdout))
	log.Info(fmt.Sprintf("starting the update for the feed @ %s", cfg.Feed.Url))
	//
	httpClient := http.Client{
		Timeout: cfg.Feed.UpdateTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.Feed.TlsSkipVerify,
			},
		},
	}
	feedsClient := feeds.NewClient(httpClient, cfg.Feed.UserAgent)
	feedsClient = feeds.NewLoggingMiddleware(feedsClient, log)
	log.Info("initialized the RSS client")
	//
	var stor feeds.Storage
	stor, err = feeds.NewStorage(ctx, cfg.Db)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the storage: %s", err))
	}
	defer stor.Close()
	//
	var feedUpdTime time.Time
	feedUpdTime, err = stor.GetUpdateTime(ctx, cfg.Feed.Url)
	if err != nil {
		panic(fmt.Sprintf("failed to read the feed update time: %s", err))
	}
	log.Info(fmt.Sprintf("feed %s: update time is %s", cfg.Feed.Url, feedUpdTime.Format(time.RFC3339)))
	//
	var awakariClient api.Client
	awakariClient, err = api.
		NewClientBuilder().
		WriteUri(cfg.Api.Writer.Uri).
		Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the Awakari API client: %s", err))
	}
	defer awakariClient.Close()
	log.Info("initialized the Awakari API client")
	//
	ws, err := awakariClient.WriteMessages(context.TODO(), "producer-rss")
	if err != nil {
		panic(fmt.Sprintf("failed to open the messages write stream: %s", err))
	}
	defer ws.Close()
	log.Info("opened the messages write stream")
	//
	var feed *rss.Feed
	feed, err = rss.FetchByFunc(feedsClient.Get, cfg.Feed.Url)
	if err != nil {
		log.Error(fmt.Sprintf("failed to fetch the feed @ %s:", cfg.Feed.Url), err)
	}
	var newFeedUpdTime time.Time
	if err == nil {
		newFeedUpdTime, err = process(cfg, feed, feedUpdTime, ws)
	}
	if err != nil {
		log.Error(fmt.Sprintf("failed to process the feed @ %s:", cfg.Feed.Url), err)
	}
	if newFeedUpdTime.After(feedUpdTime) {
		err = stor.SetUpdateTime(ctx, cfg.Feed.Url, newFeedUpdTime)
	}
	if err != nil {
		log.Error(fmt.Sprintf("failed to set the new update time for the feed @ %s:", cfg.Feed.Url), err)
	}
}

func process(cfg config.Config, feed *rss.Feed, prevTime time.Time, ws model.WriteStream[*pb.CloudEvent]) (nextTime time.Time, err error) {
	var msgs []*pb.CloudEvent
	msgs, nextTime, err = convertNewItemsToMessages(cfg.Message, feed, prevTime)
	if err == nil && len(msgs) > 0 {
		err = sendMessages(msgs, ws, cfg.Api.Writer.Backoff)
	}
	return
}

func convertNewItemsToMessages(msgCfg config.MessageConfig, feed *rss.Feed, prevTime time.Time) (msgs []*pb.CloudEvent, nextTime time.Time, err error) {
	var msg *pb.CloudEvent
	var itemErr error
	for _, item := range feed.Items {
		if prevTime.Before(item.Date) {
			msg, itemErr = convertToMessage(msgCfg, feed, item)
			msgs = append(msgs, msg)
			err = errors.Join(err, itemErr)
		}
		if nextTime.Before(item.Date) {
			nextTime = item.Date
		}
	}
	return
}

func convertToMessage(msgCfg config.MessageConfig, feed *rss.Feed, item *rss.Item) (msg *pb.CloudEvent, err error) {
	attrs := map[string]*pb.CloudEventAttributeValue{
		"subject": {
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: item.Link,
			},
		},
		"time": {
			Attr: &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(item.Date),
			},
		},
	}
	if feed.Author != "" {
		attrs[msgCfg.Metadata.KeyAuthor] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: feed.Author,
			},
		}
	}
	if len(feed.Categories) > 0 {
		attrs[msgCfg.Metadata.KeyFeedCategories] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: strings.Join(feed.Categories, " "),
			},
		}
	}
	if feed.Description != "" {
		attrs[msgCfg.Metadata.KeyFeedDescription] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: feed.Description,
			},
		}
	}
	if feed.Image != nil {
		if feed.Image.Title != "" {
			attrs[msgCfg.Metadata.KeyFeedImageTitle] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: feed.Image.Title,
				},
			}
		}
		if feed.Image.URL != "" {
			attrs[msgCfg.Metadata.KeyFeedImageUrl] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeUri{
					CeUri: feed.Image.URL,
				},
			}
		}
	}
	if feed.Language != "" {
		attrs[msgCfg.Metadata.KeyLanguage] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: feed.Language,
			},
		}
	}
	if feed.Title != "" {
		attrs[msgCfg.Metadata.KeyFeedTitle] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: feed.Title,
			},
		}
	}
	if item.ID != "" {
		attrs[msgCfg.Metadata.KeyGuid] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: item.ID,
			},
		}
	}
	if len(item.Categories) > 0 {
		attrs[msgCfg.Metadata.KeyCategories] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: strings.Join(item.Categories, " "),
			},
		}
	}
	if item.Image == nil {
		for _, encl := range item.Enclosures {
			if strings.HasPrefix(encl.Type, "image/") {
				if encl.URL != "" {
					attrs[msgCfg.Metadata.KeyImageUrl] = &pb.CloudEventAttributeValue{
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: encl.URL,
						},
					}
				}
				break
			}
		}
	} else {
		if item.Image.Title != "" {
			attrs[msgCfg.Metadata.KeyImageTitle] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: item.Image.Title,
				},
			}
		}
		if item.Image.URL != "" {
			attrs[msgCfg.Metadata.KeyImageUrl] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: item.Image.URL,
				},
			}
		}
	}
	if item.Summary != "" {
		attrs[msgCfg.Metadata.KeySummary] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: item.Summary,
			},
		}
	}
	if item.Title != "" {
		attrs[msgCfg.Metadata.KeyTitle] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: item.Title,
			},
		}
	}
	msg = &pb.CloudEvent{
		Id:          uuid.NewString(),
		SpecVersion: msgCfg.Metadata.SpecVersion,
		Source:      feed.Link,
		Type:        "com.github.awakari.producer-rss",
		Attributes:  attrs,
	}
	if item.Content != "" {
		msg.Data = &pb.CloudEvent_TextData{
			TextData: item.Content,
		}
	}
	return
}

func sendMessages(msgs []*pb.CloudEvent, ws model.WriteStream[*pb.CloudEvent], clientBackoff time.Duration) (err error) {
	msgCount := uint32(len(msgs))
	var ackCount, n uint32
	for ackCount < msgCount {
		n, err = ws.WriteBatch(msgs[ackCount:])
		ackCount += n
		if err != nil {
			break
		}
		if n == 0 {
			time.Sleep(clientBackoff)
		}
	}
	return
}
