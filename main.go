package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/SlyMarbo/rss"
	"github.com/awakari/client-sdk-go/api"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"producer-rss/config"
	"producer-rss/feeds"
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
	//
	conv := NewConverter(cfg.Message)
	conv = NewConverterLogging(conv, log)
	prod := NewProducer(feed, feedUpdTime, conv, ws, cfg.Api.Writer.Backoff)
	prod = NewProducerLogging(prod, log)
	//
	var newFeedUpdTime time.Time
	if err == nil {
		newFeedUpdTime, err = prod.Produce(ctx)
	}
	if err != nil {
		log.Error(fmt.Sprintf("failed to process the feed @ %s:", cfg.Feed.Url), err)
	}
	if newFeedUpdTime.After(feedUpdTime) {
		log.Info(fmt.Sprintf("setting the new update time to %s", newFeedUpdTime.Format(time.RFC3339)))
		err = stor.SetUpdateTime(ctx, cfg.Feed.Url, newFeedUpdTime)
	}
	if err != nil {
		log.Error(fmt.Sprintf("failed to set the new update time for the feed @ %s:", cfg.Feed.Url), err)
	}
}
