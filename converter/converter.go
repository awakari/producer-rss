package converter

import (
	"github.com/SlyMarbo/rss"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"producer-rss/config"
	"strings"
)

type Converter interface {
	Convert(feed *rss.Feed, item *rss.Item) (msg *pb.CloudEvent)
}

type converter struct {
	cfgMsg config.MessageConfig
}

func NewConverter(cfgMsg config.MessageConfig) Converter {
	return converter{
		cfgMsg: cfgMsg,
	}
}

func (c converter) Convert(feed *rss.Feed, item *rss.Item) (msg *pb.CloudEvent) {
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
		attrs[c.cfgMsg.Metadata.KeyAuthor] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: feed.Author,
			},
		}
	}
	if len(feed.Categories) > 0 {
		attrs[c.cfgMsg.Metadata.KeyFeedCategories] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: strings.Join(feed.Categories, " "),
			},
		}
	}
	if feed.Description != "" {
		attrs[c.cfgMsg.Metadata.KeyFeedDescription] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: feed.Description,
			},
		}
	}
	if feed.Image != nil {
		if feed.Image.Title != "" {
			attrs[c.cfgMsg.Metadata.KeyFeedImageTitle] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: feed.Image.Title,
				},
			}
		}
		if feed.Image.URL != "" {
			attrs[c.cfgMsg.Metadata.KeyFeedImageUrl] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeUri{
					CeUri: feed.Image.URL,
				},
			}
		}
	}
	if feed.Language != "" {
		attrs[c.cfgMsg.Metadata.KeyLanguage] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: feed.Language,
			},
		}
	}
	if feed.Title != "" {
		attrs[c.cfgMsg.Metadata.KeyFeedTitle] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: feed.Title,
			},
		}
	}
	if item.ID != "" {
		attrs[c.cfgMsg.Metadata.KeyGuid] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: item.ID,
			},
		}
	}
	if len(item.Categories) > 0 {
		attrs[c.cfgMsg.Metadata.KeyCategories] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: strings.Join(item.Categories, " "),
			},
		}
	}
	if item.Image == nil {
		for _, encl := range item.Enclosures {
			if strings.HasPrefix(encl.Type, "image/") {
				if encl.URL != "" {
					attrs[c.cfgMsg.Metadata.KeyImageUrl] = &pb.CloudEventAttributeValue{
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
			attrs[c.cfgMsg.Metadata.KeyImageTitle] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: item.Image.Title,
				},
			}
		}
		if item.Image.URL != "" {
			attrs[c.cfgMsg.Metadata.KeyImageUrl] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: item.Image.URL,
				},
			}
		}
	}
	if item.Summary != "" {
		attrs[c.cfgMsg.Metadata.KeySummary] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: item.Summary,
			},
		}
	}
	if item.Title != "" {
		attrs[c.cfgMsg.Metadata.KeyTitle] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: item.Title,
			},
		}
	}
	msg = &pb.CloudEvent{
		Id:          uuid.NewString(),
		SpecVersion: c.cfgMsg.Metadata.SpecVersion,
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
