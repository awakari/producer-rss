package test

import (
	"github.com/SlyMarbo/rss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_Parse(t *testing.T) {
	cases := map[string]struct {
		count int
	}{
		"data/export-arxiv-org-rss-astro-ph.xml": {
			count: 118,
		},
		"data/rss-cnn-com-rss-edition-rss.xml": {
			count: 50,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			data, err := os.ReadFile(k)
			require.Nil(t, err)
			feed, err := rss.Parse(data)
			assert.Equal(t, c.count, len(feed.Items))
			assert.Nil(t, err)
		})
	}
}
