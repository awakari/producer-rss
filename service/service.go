package service

import (
	"context"
	"errors"
	"github.com/SlyMarbo/rss"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
	"net/url"
	"producer-rss/api/grpc/resolver"
	"producer-rss/config"
	"strings"
	"time"
)

type Service interface {
	ProcessLoop(errChan chan<- error)
}

type service struct {
	feeds             []*rss.Feed
	feedsClient       Client
	updateIntervalMin time.Duration
	updateIntervalMax time.Duration
	msgCfg            config.MessageConfig
	resolverBackoff   time.Duration
	resolverSvc       resolver.Service
}

func NewService(feedCfg config.FeedConfig, feedsClient Client, msgCfg config.MessageConfig, resolverBackoff time.Duration, resolverSvc resolver.Service) (svc Service, err error) {
	var feeds []*rss.Feed
	for _, feedUrl := range feedCfg.Urls {
		feed, fetchErr := rss.FetchByFunc(feedsClient.Get, feedUrl)
		if fetchErr != nil {
			err = errors.Join(err, fetchErr)
		} else {
			if feed.Refresh.Sub(time.Now()) > feedCfg.UpdateIntervalMax {
				feed.Refresh = time.Now().Add(feedCfg.UpdateIntervalMax)
			}
			feeds = append(feeds, feed)
		}
	}
	svc = service{
		feeds:             feeds,
		feedsClient:       feedsClient,
		updateIntervalMin: feedCfg.UpdateIntervalMin,
		updateIntervalMax: feedCfg.UpdateIntervalMax,
		msgCfg:            msgCfg,
		resolverBackoff:   resolverBackoff,
		resolverSvc:       resolverSvc,
	}
	return
}

func (svc service) ProcessLoop(errChan chan<- error) {
	for {
		err := svc.ProcessFeeds()
		if err != nil {
			errChan <- err
		}
		time.Sleep(svc.updateIntervalMin)
	}
}

func (svc service) ProcessFeeds() (err error) {
	for _, feed := range svc.feeds {
		nextErr := svc.processFeed(feed)
		if nextErr != nil {
			err = errors.Join(err, nextErr)
		}
	}
	return
}

func (svc service) processFeed(feed *rss.Feed) (err error) {
	if feed.Refresh.Before(time.Now()) {
		err = feed.UpdateByFunc(svc.feedsClient.Get)
		if feed.Refresh.Sub(time.Now()) > svc.updateIntervalMax {
			feed.Refresh = time.Now().Add(svc.updateIntervalMax)
		}
		if err == nil {
			var msgs []*event.Event
			msgs, err = svc.convertToMessages(feed)
			if err == nil && len(msgs) > 0 {
				err = svc.processMessages(msgs)
			}
		}
	}
	return
}

func (svc service) convertToMessages(feed *rss.Feed) (msgs []*event.Event, err error) {
	var msg *event.Event
	for _, item := range feed.Items {
		if !item.Read {
			item.Read = true
			msg, err = svc.convertToMessage(feed, item)
			msgs = append(msgs, msg)
		}
	}
	return
}

func (svc service) processMessages(msgs []*event.Event) (err error) {
	msgCount := uint32(len(msgs))
	var ackCount, n uint32
	for ackCount < msgCount {
		n, err = svc.resolverSvc.SubmitBatch(context.TODO(), msgs[ackCount:])
		ackCount += n
		if err != nil {
			break
		}
		if n == 0 {
			time.Sleep(svc.resolverBackoff)
		}
	}
	return
}

func (svc service) convertToMessage(feed *rss.Feed, item *rss.Item) (msg *event.Event, err error) {
	msg = &event.Event{}
	msg.SetSpecVersion(svc.msgCfg.Metadata.SpecVersion)
	msg.SetID(uuid.NewString())
	msg.SetSource(feed.Link)
	msg.SetSubject(item.Link)
	msg.SetTime(item.Date)
	if feed.Author != "" {
		msg.SetExtension(svc.msgCfg.Metadata.KeyAuthor, feed.Author)
	}
	if len(feed.Categories) > 0 {
		msg.SetExtension(svc.msgCfg.Metadata.KeyFeedCategories, strings.Join(feed.Categories, " "))
	}
	if feed.Description != "" {
		msg.SetExtension(svc.msgCfg.Metadata.KeyFeedDescription, feed.Description)
	}
	if feed.Image != nil {
		if feed.Image.Title != "" {
			msg.SetExtension(svc.msgCfg.Metadata.KeyFeedImageTitle, feed.Image.Title)
		}
		if feed.Image.URL != "" {
			feedImgUrl, _ := url.Parse(feed.Image.URL)
			if feedImgUrl != nil {
				msg.SetExtension(svc.msgCfg.Metadata.KeyFeedImageUrl, feedImgUrl)
			}
		}
	}
	if feed.Language != "" {
		msg.SetExtension(svc.msgCfg.Metadata.KeyLanguage, feed.Language)
	}
	if feed.Title != "" {
		msg.SetExtension(svc.msgCfg.Metadata.KeyFeedTitle, feed.Title)
	}
	if item.ID != "" {
		msg.SetExtension(svc.msgCfg.Metadata.KeyGuid, item.ID)
	}
	if len(item.Categories) > 0 {
		msg.SetExtension(svc.msgCfg.Metadata.KeyCategories, strings.Join(item.Categories, " "))
	}
	if item.Image == nil {
		for _, encl := range item.Enclosures {
			if strings.HasPrefix(encl.Type, "image/") {
				itemImgUrl, _ := url.Parse(encl.URL)
				if itemImgUrl != nil {
					msg.SetExtension(svc.msgCfg.Metadata.KeyImageUrl, itemImgUrl)
				}
				break
			}
		}
	} else {
		if item.Image.Title != "" {
			msg.SetExtension(svc.msgCfg.Metadata.KeyImageTitle, item.Image.Title)
		}
		if item.Image.URL != "" {
			itemImgUrl, _ := url.Parse(item.Image.URL)
			if itemImgUrl != nil {
				msg.SetExtension(svc.msgCfg.Metadata.KeyImageUrl, itemImgUrl)
			}
		}
	}
	if item.Summary != "" {
		msg.SetExtension(svc.msgCfg.Metadata.KeySummary, item.Summary)
	}
	if item.Title != "" {
		msg.SetExtension(svc.msgCfg.Metadata.KeyTitle, item.Title)
	}
	if item.Content != "" {
		err = msg.SetData(svc.msgCfg.Content.Type, item.Content)
	}
	return
}
