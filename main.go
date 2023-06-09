package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/SlyMarbo/rss"
	"github.com/awakari/client-sdk-go/api"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc/metadata"
	"net/http"
	"os"
	"producer-rss/config"
	"producer-rss/converter"
	"producer-rss/feeds"
	"producer-rss/producer"
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
		WriterUri(cfg.Api.Writer.Uri).
		Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the Awakari API client: %s", err))
	}
	defer awakariClient.Close()
	log.Info("initialized the Awakari API client")
	//
	groupIdCtx := metadata.AppendToOutgoingContext(
		ctx,
		"x-awakari-group-id", "producer-rss",
		"x-awakari-user-id", "producer-rss",
	)
	ws, err := awakariClient.OpenMessagesWriter(groupIdCtx, "producer-rss")
	if err != nil {
		panic(fmt.Sprintf("failed to open the messages writer: %s", err))
	}
	defer ws.Close()
	log.Info("opened the messages writer")
	//
	var feed *rss.Feed
	feed, err = rss.FetchByFunc(feedsClient.Get, cfg.Feed.Url)
	if err != nil {
		log.Error(fmt.Sprintf("failed to fetch the feed: %s:", err))
	}
	log.Info(fmt.Sprintf("feed contains %d items to process", len(feed.Items)))
	//
	conv := converter.NewConverter(cfg.Message)
	conv = converter.NewConverterLogging(conv, log)
	prod := producer.NewProducer(feed, feedUpdTime, conv, ws, cfg.Api.Writer.Backoff, cfg.Api.Writer.BatchSize)
	prod = producer.NewProducerLogging(prod, log)
	//
	var newFeedUpdTime time.Time
	if err == nil {
		newFeedUpdTime, err = prod.Produce(ctx)
	}
	if err != nil {
		log.Error(fmt.Sprintf("failed to process the feed: %s", err))
	}
	if newFeedUpdTime.After(feedUpdTime) {
		log.Info(fmt.Sprintf("setting the new update time to %s", newFeedUpdTime.Format(time.RFC3339)))
		err = stor.SetUpdateTime(ctx, cfg.Feed.Url, newFeedUpdTime)
	}
	if err != nil {
		log.Error(fmt.Sprintf("failed to set the new update time for the feed: %s:", err))
	}
}
