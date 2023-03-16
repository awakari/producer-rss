package config

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"os"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	os.Setenv("API_RESOLVER_BACKOFF", "23h")
	os.Setenv("API_RESOLVER_URI", "resolver:1234")
	os.Setenv("LOG_LEVEL", "4")
	os.Setenv("FEED_UPDATE_TIMEOUT", "34ms")
	os.Setenv("MSG_MD_KEY_FEED_TITLE", "feed title")
	os.Setenv("MSG_MD_KEY_LANGUAGE", "lang")
	os.Setenv("MSG_CONTENT_TYPE", "text/xml")
	cfg, err := NewConfigFromEnv()
	assert.Nil(t, err)
	assert.Equal(t, 23*time.Hour, cfg.Api.Resolver.Backoff)
	assert.Equal(t, "resolver:1234", cfg.Api.Resolver.Uri)
	assert.Equal(t, slog.LevelWarn, slog.Level(cfg.Log.Level))
	assert.Equal(t, 34*time.Millisecond, cfg.Feed.UpdateTimeout)
	assert.Equal(t, "feed title", cfg.Message.Metadata.KeyFeedTitle)
	assert.Equal(t, "lang", cfg.Message.Metadata.KeyLanguage)
	assert.Equal(t, "text/xml", cfg.Message.Content.Type)
}
