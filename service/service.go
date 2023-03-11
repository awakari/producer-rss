package service

import (
	"context"
	"github.com/SlyMarbo/rss"
	"github.com/cloudevents/sdk-go/v2/event"
	"producer-rss/api/grpc/resolver"
	"producer-rss/config"
	"time"
)

type Service interface {
	Start()
}

type service struct {
	feeds          []*rss.Feed
	updateInterval time.Duration
	msgMdCfg       config.MessageMetadataConfig
	resolverSvc    resolver.Service
}

func NewService(feedUris []string, feedCfg config.FeedConfig, msgMdCfg config.MessageMetadataConfig, resolverSvc resolver.Service) (svc Service, err error) {
	var feed *rss.Feed
	var feeds []*rss.Feed
	updateInterval := feedCfg.UpdateInterval.Max
	now := time.Now()
	for _, feedUri := range feedUris {
		feed, err = rss.Fetch(feedUri)
		if err != nil {
			break
		}
		timeUntilUpdate := feed.Refresh.Sub(now)
		if timeUntilUpdate < updateInterval {
			if timeUntilUpdate < feedCfg.UpdateInterval.Min {
				updateInterval = feedCfg.UpdateInterval.Min
			} else {
				updateInterval = timeUntilUpdate
			}
		}
		feeds = append(feeds, feed)
	}
	rss.DefaultRefreshInterval = updateInterval
	if err == nil {
		svc = service{
			feeds:          feeds,
			updateInterval: updateInterval,
			msgMdCfg:       msgMdCfg,
			resolverSvc:    resolverSvc,
		}
	}
	return svc, err
}

func (svc service) Start() {
	go svc.processLoop()
}

func (svc service) processLoop() {
	for {
		for _, feed := range svc.feeds {
			svc.processFeed(feed)
		}
		time.Sleep(svc.updateInterval)
	}
}

func (svc service) processFeed(feed *rss.Feed) {
	if feed.Refresh.Before(time.Now()) {
		err := feed.Update()
		if err == nil {
			msgs := convertToMessages(feed)
			processMessages(svc.resolverSvc, msgs)
		}
	}
}

func convertToMessages(feed *rss.Feed) (msgs []*event.Event) {
	for _, item := range feed.Items {
		msg := convertToMessage(feed, item)
		msgs = append(msgs, msg)
	}
	return
}

func convertToMessage(feed *rss.Feed, item *rss.Item) (msg *event.Event) {
	msg := event.New()
	msg.SetID(item.ID)
	msg.SetSource(feed.Link)
	msg.SetSubject(item.Link)
	msg.SetTime(item.Date)
	msg.SetSpecVersion()
	feed.Author
	msg.SetExtension("categories", item.Categories) // TODO join w/ space
	msg.SetExtension("imagetitle", item.Image.Title)
	msg.SetExtension("imageurl", item.Image.URL)
	msg.SetExtension("summary", item.Summary)
	msg.SetExtension("title", item.Title)
	msg.SetData("text/plain", item.Content)
	return
}

func processMessages(resolverSvc resolver.Service, msgs []*event.Event) {
	for _, msg := range msgs {
		_ = resolverSvc.Submit(context.TODO(), msg)
	}
}
