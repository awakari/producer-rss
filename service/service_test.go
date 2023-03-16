package service

import (
	"github.com/SlyMarbo/rss"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"producer-rss/api/grpc/resolver"
	"producer-rss/config"
	"testing"
	"time"
)

func TestService_ProcessFeeds(t *testing.T) {
	cfg, err := config.NewConfigFromEnv()
	cfg.Feed.Urls = []string{
		"http://test.rss",
	}
	require.Nil(t, err)
	feedsClient := NewClientMock()
	resolverSvc := resolver.NewServiceMock()
	svc, err := NewService(cfg.Feed, feedsClient, cfg.Message, cfg.Api.Resolver.Backoff, resolverSvc)
	require.Nil(t, err)
	err = svc.(service).ProcessFeeds()
	assert.Nil(t, err)
}

func TestService_convertToMessages(t *testing.T) {
	cfg, err := config.NewConfigFromEnv()
	cfg.Feed.Urls = []string{
		"http://test.rss",
	}
	require.Nil(t, err)
	feedsClient := NewClientMock()
	resolverSvc := resolver.NewServiceMock()
	svc, err := NewService(cfg.Feed, feedsClient, cfg.Message, cfg.Api.Resolver.Backoff, resolverSvc)
	require.Nil(t, err)
	dt, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	require.Nil(t, err)
	feed := &rss.Feed{
		Title:       "feed title",
		Language:    "feed lang",
		Author:      "feed author",
		Description: "feed description",
		Link:        "http://test.rss",
		UpdateURL:   "http://test.rss",
		Image: &rss.Image{
			Title: "img0",
			URL:   "http://img0.foo",
		},
		Categories: []string{
			"cat0",
			"cat1",
		},
		Items: []*rss.Item{
			{
				Title:   "item title 0",
				Summary: "item summary 0",
				Content: "item content 0",
				Categories: []string{
					"cat1",
					"cat2",
				},
				Link: "http://test.rss/item0",
				Date: dt,
				Image: &rss.Image{
					Title: "img1",
					URL:   "http://img1.foo",
				},
				ID: "http://test.rss/item0",
			},
			{
				Title:   "item title 1",
				Summary: "item summary 1",
				Content: "item content 1",
				Categories: []string{
					"cat3",
				},
				Link: "http://test.rss/item1",
				Date: dt,
				ID:   "http://test.rss/item1",
				Enclosures: []*rss.Enclosure{
					{
						URL:  "http://enclosure0.png",
						Type: "image/png",
					},
				},
			},
		},
		Refresh: dt,
	}
	msgs, err := svc.(service).convertToMessages(feed)
	assert.Nil(t, err)
	assert.Len(t, msgs, 2)
	//
	assert.Equal(t, "http://test.rss", msgs[0].Source())
	assert.Equal(t, []byte("item content 0"), msgs[0].Data())
	assert.Equal(t, "", msgs[0].Type())
	assert.Equal(t, time.Time(time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC)), msgs[0].Time())
	assert.Equal(t, "http://test.rss/item0", msgs[0].Subject())
	assert.Equal(t, "", msgs[0].DataSchema())
	assert.Equal(t, "text/plain", msgs[0].DataContentType())
	assert.Equal(t, "text/plain", msgs[0].DataMediaType())
	assert.Equal(t, "cat0 cat1", msgs[0].Extensions()["feedcategories"])
	assert.Equal(t, "feed description", msgs[0].Extensions()["feeddescription"])
	assert.Equal(t, "feed title", msgs[0].Extensions()["feedtitle"])
	u := msgs[0].Extensions()["feedimageurl"].(types.URI)
	assert.Equal(t, "http://img0.foo", u.String())
	assert.Equal(t, "img0", msgs[0].Extensions()["feedimagetitle"])
	assert.Equal(t, "feed author", msgs[0].Extensions()["author"])
	assert.Equal(t, "cat1 cat2", msgs[0].Extensions()["categories"])
	assert.Equal(t, "feed lang", msgs[0].Extensions()["language"])
	assert.Equal(t, "http://test.rss/item0", msgs[0].Extensions()["rssitemguid"])
	assert.Equal(t, "img1", msgs[0].Extensions()["imagetitle"])
	u = msgs[0].Extensions()["imageurl"].(types.URI)
	assert.Equal(t, "http://img1.foo", u.String())
	assert.Equal(t, "item summary 0", msgs[0].Extensions()["summary"])
	assert.Equal(t, "item title 0", msgs[0].Extensions()["title"])
	//
	assert.Equal(t, "http://test.rss", msgs[1].Source())
	assert.Equal(t, []byte("item content 1"), msgs[1].Data())
	assert.Equal(t, "", msgs[1].Type())
	assert.Equal(t, time.Time(time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC)), msgs[1].Time())
	assert.Equal(t, "http://test.rss/item1", msgs[1].Subject())
	assert.Equal(t, "", msgs[1].DataSchema())
	assert.Equal(t, "text/plain", msgs[1].DataContentType())
	assert.Equal(t, "text/plain", msgs[1].DataMediaType())
	assert.Equal(t, "cat0 cat1", msgs[1].Extensions()["feedcategories"])
	assert.Equal(t, "feed description", msgs[1].Extensions()["feeddescription"])
	assert.Equal(t, "feed title", msgs[1].Extensions()["feedtitle"])
	u = msgs[1].Extensions()["feedimageurl"].(types.URI)
	assert.Equal(t, "http://img0.foo", u.String())
	assert.Equal(t, "img0", msgs[1].Extensions()["feedimagetitle"])
	assert.Equal(t, "feed author", msgs[1].Extensions()["author"])
	assert.Equal(t, "cat3", msgs[1].Extensions()["categories"])
	assert.Equal(t, "feed lang", msgs[1].Extensions()["language"])
	assert.Equal(t, "http://test.rss/item1", msgs[1].Extensions()["rssitemguid"])
	u = msgs[1].Extensions()["imageurl"].(types.URI)
	assert.Equal(t, "http://enclosure0.png", u.String())
	assert.Equal(t, "item summary 1", msgs[1].Extensions()["summary"])
	assert.Equal(t, "item title 1", msgs[1].Extensions()["title"])
}

func TestService_processFeed(t *testing.T) {
	cfg, err := config.NewConfigFromEnv()
	cfg.Feed.Urls = []string{
		"http://test.rss",
	}
	require.Nil(t, err)
	feedsClient := NewClientMock()
	resolverSvc := resolver.NewServiceMock()
	svc, err := NewService(cfg.Feed, feedsClient, cfg.Message, cfg.Api.Resolver.Backoff, resolverSvc)
	require.Nil(t, err)
	dt, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	require.Nil(t, err)
	feed := &rss.Feed{
		Title:       "feed title",
		Language:    "feed lang",
		Author:      "feed author",
		Description: "feed description",
		Link:        "http://test.rss",
		UpdateURL:   "http://test.rss",
		Image: &rss.Image{
			Title: "img0",
			URL:   "http://img0.foo",
		},
		Categories: []string{
			"cat0",
			"cat1",
		},
		Items: []*rss.Item{
			{
				Title:   "item title 0",
				Summary: "item summary 0",
				Content: "item content 0",
				Categories: []string{
					"cat1",
					"cat2",
				},
				Link: "http://test.rss/item0",
				Date: dt,
				Image: &rss.Image{
					Title: "img1",
					URL:   "http://img1.foo",
				},
				ID: "http://test.rss/item0",
			},
			{
				Title:   "item title 1",
				Summary: "item summary 1",
				Content: "item content 1",
				Categories: []string{
					"cat3",
				},
				Link: "http://test.rss/item1",
				Date: dt,
				ID:   "http://test.rss/item1",
				Enclosures: []*rss.Enclosure{
					{
						URL:  "http://enclosure0.png",
						Type: "image/png",
					},
				},
			},
		},
		Refresh: dt,
	}
	err = svc.(service).processFeed(feed)
	assert.Nil(t, err)
}
