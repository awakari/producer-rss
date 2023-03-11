package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	Api struct {
		Resolver struct {
			Uri string `envconfig:"API_RESOLVER_URI" default:"resolver:8080" required:"true"`
		}
	}
	Feed    FeedConfig
	Message struct {
		Metadata MessageMetadataConfig
	}
}

type FeedConfig struct {
	UpdateInterval struct {
		Min time.Duration `envconfig:"FEED_UPDATE_INTERVAL_MIN" default:"10s" required:"true"`
		Max time.Duration `envconfig:"FEED_UPDATE_INTERVAL_MAX" default:"24h" required:"true"`
	}
}

type MessageMetadataConfig struct {
	//
	KeyFeedCategories  string `envconfig:"MSG_MD_KEY_FEED_CATEGORIES" default:"feedcategories" required:"true"`
	KeyFeedDescription string `envconfig:"MSG_MD_KEY_FEED_DESCRIPTION" default:"feeddescription" required:"true"`
	KeyFeedImageTitle  string `envconfig:"MSG_MD_KEY_FEED_IMAGE_TITLE" default:"feedimagetitle" required:"true"`
	KeyFeedImageUrl    string `envconfig:"MSG_MD_KEY_FEED_IMAGE_URL" default:"feedimageurl" required:"true"`
	KeyFeedTitle       string `envconfig:"MSG_MD_KEY_FEED_TITLE" default:"feedtitle" required:"true"`
	//
	KeyAuthor     string `envconfig:"MSG_MD_KEY_AUTHOR" default:"author" required:"true"`
	KeyCategories string `envconfig:"MSG_MD_KEY_CATEGORIES" default:"categories" required:"true"`
	KeyImageTitle string `envconfig:"MSG_MD_KEY_IMAGE_TITLE" default:"imagetitle" required:"true"`
	KeyImageUrl   string `envconfig:"MSG_MD_KEY_IMAGE_URL" default:"imageurl" required:"true"`
	KeyLanguage   string `envconfig:"MSG_MD_KEY_LANGUAGE" default:"language" required:"true"`
	KeySummary    string `envconfig:"MSG_MD_KEY_SUMMARY" default:"summary" required:"true"`
	KeyTitle      string `envconfig:"MSG_MD_KEY_TITLE" default:"title" required:"true"`
	//
	SpecVersion string `envconfig:"MSG_MD_SPEC_VERSION" default:"1.0" required:"true"`
}

func NewConfigFromEnv() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
