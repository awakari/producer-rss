package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	Api struct {
		Writer struct {
			Backoff time.Duration `envconfig:"API_WRITER_BACKOFF" default:"10s" required:"true"`
			Uri     string        `envconfig:"API_WRITER_URI" default:"writer:50051" required:"true"`
		}
	}
	Db   DbConfig
	Feed FeedConfig
	Log  struct {
		Level int `envconfig:"LOG_LEVEL" default:"-4" required:"true"`
	}
	Message MessageConfig
}

type DbConfig struct {
	Uri      string `envconfig:"DB_URI" default:"mongodb://localhost:27017/?retryWrites=true&w=majority" required:"true"`
	Name     string `envconfig:"DB_NAME" default:"producer-rss" required:"true"`
	UserName string `envconfig:"DB_USERNAME" default:""`
	Password string `envconfig:"DB_PASSWORD" default:""`
	Table    struct {
		Name string `envconfig:"DB_TABLE_NAME" default:"feeds" required:"true"`
	}
	Tls struct {
		Enabled  bool `envconfig:"DB_TLS_ENABLED" default:"false" required:"true"`
		Insecure bool `envconfig:"DB_TLS_INSECURE" default:"false" required:"true"`
	}
}

type FeedConfig struct {
	Url               string        `envconfig:"FEED_URL" required:"true"`
	TlsSkipVerify     bool          `envconfig:"FEED_TLS_SKIP_VERIFY" default:"true" required:"true"`
	UpdateIntervalMin time.Duration `envconfig:"FEED_UPDATE_INTERVAL_MIN" default:"10s" required:"true"`
	UpdateIntervalMax time.Duration `envconfig:"FEED_UPDATE_INTERVAL_MAX" default:"10m" required:"true"`
	UpdateTimeout     time.Duration `envconfig:"FEED_UPDATE_TIMEOUT" default:"1m" required:"true"`
	UserAgent         string        `envconfig:"FEED_USER_AGENT" default:"awakari-producer-rss/0.0.1" required:"true"`
}

type MessageConfig struct {
	Metadata MetadataConfig
	Content  ContentConfig
}

type MetadataConfig struct {
	//
	KeyFeedCategories  string `envconfig:"MSG_MD_KEY_FEED_CATEGORIES" default:"feedcategories" required:"true"`
	KeyFeedDescription string `envconfig:"MSG_MD_KEY_FEED_DESCRIPTION" default:"feeddescription" required:"true"`
	KeyFeedImageTitle  string `envconfig:"MSG_MD_KEY_FEED_IMAGE_TITLE" default:"feedimagetitle" required:"true"`
	KeyFeedImageUrl    string `envconfig:"MSG_MD_KEY_FEED_IMAGE_URL" default:"feedimageurl" required:"true"`
	KeyFeedTitle       string `envconfig:"MSG_MD_KEY_FEED_TITLE" default:"feedtitle" required:"true"`
	//
	KeyAuthor     string `envconfig:"MSG_MD_KEY_AUTHOR" default:"author" required:"true"`
	KeyCategories string `envconfig:"MSG_MD_KEY_CATEGORIES" default:"categories" required:"true"`
	KeyGuid       string `envconfig:"MSG_MD_KEY_GUID" default:"rssitemguid" required:"true"`
	KeyImageTitle string `envconfig:"MSG_MD_KEY_IMAGE_TITLE" default:"imagetitle" required:"true"`
	KeyImageUrl   string `envconfig:"MSG_MD_KEY_IMAGE_URL" default:"imageurl" required:"true"`
	KeyLanguage   string `envconfig:"MSG_MD_KEY_LANGUAGE" default:"language" required:"true"`
	KeySummary    string `envconfig:"MSG_MD_KEY_SUMMARY" default:"summary" required:"true"`
	KeyTitle      string `envconfig:"MSG_MD_KEY_TITLE" default:"title" required:"true"`
	//
	SpecVersion string `envconfig:"MSG_MD_SPEC_VERSION" default:"1.0" required:"true"`
}

type ContentConfig struct {
	Type string `envconfig:"MSG_CONTENT_TYPE" default:"text/plain" required:"true"`
}

func NewConfigFromEnv() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
