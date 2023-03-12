package service

import (
	"context"
	"errors"
	"github.com/SlyMarbo/rss"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
	"net/http"
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
	fetchFunc         func(url string) (resp *http.Response, err error)
	updateIntervalMin time.Duration
	msgCfg            config.MessageConfig
	resolverSvc       resolver.Service
}

func NewService(feedCfg config.FeedConfig, msgCfg config.MessageConfig, resolverSvc resolver.Service) (svc Service, err error) {
	var feeds []*rss.Feed
	rss.DefaultRefreshInterval = feedCfg.UpdateIntervalMax
	fetchFunc := func(url string) (resp *http.Response, err error) {
		return fetch(url, feedCfg.UserAgent)
	}
	for _, feedUrl := range feedCfg.Urls {
		feed, fetchErr := rss.FetchByFunc(fetchFunc, feedUrl)
		if fetchErr != nil {
			err = errors.Join(err, fetchErr)
		} else {
			feeds = append(feeds, feed)
		}
	}
	if err == nil {
		svc = service{
			feeds:             feeds,
			fetchFunc:         fetchFunc,
			updateIntervalMin: feedCfg.UpdateIntervalMin,
			msgCfg:            msgCfg,
			resolverSvc:       resolverSvc,
		}
	}
	return svc, err
}

func fetch(url, userAgent string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return http.DefaultClient.Do(req)
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
		err = feed.UpdateByFunc(svc.fetchFunc)
		if err == nil {
			msgs := svc.convertToMessages(feed)
			err = processMessages(svc.resolverSvc, msgs)
		}
	}
	return
}

func (svc service) convertToMessages(feed *rss.Feed) (msgs []*event.Event) {
	for _, item := range feed.Items {
		if !item.Read {
			item.Read = true
			msg := svc.convertToMessage(feed, item)
			msgs = append(msgs, msg)
		}
	}
	return
}

func (svc service) convertToMessage(feed *rss.Feed, item *rss.Item) *event.Event {
	msg := event.New()
	msg.SetID(uuid.NewString())
	msg.SetSource(feed.Link)
	msg.SetSubject(item.Link)
	msg.SetTime(item.Date)
	msg.SetSpecVersion(svc.msgCfg.Metadata.SpecVersion)
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
		_ = msg.SetData(svc.msgCfg.Content.Type, item.Summary)
	}
	if item.Title != "" {
		msg.SetExtension(svc.msgCfg.Metadata.KeyTitle, item.Title)
	}
	return &msg
}

func processMessages(resolverSvc resolver.Service, msgs []*event.Event) (err error) {
	for _, msg := range msgs {
		nextErr := processMessage(resolverSvc, msg)
		if nextErr != nil {
			err = errors.Join(err, nextErr)
		}
	}
	return
}

func processMessage(resolverSvc resolver.Service, msg *event.Event) (err error) {
	for {
		err = resolverSvc.Submit(context.TODO(), msg)
		if errors.Is(err, resolver.ErrQueueFull) {

		} else {
			break
		}
	}
	return
}
