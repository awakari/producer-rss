package feeds

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"producer-rss/config"
	"testing"
	"time"
)

var dbUri = os.Getenv("DB_URI_TEST_MONGO")

func TestNewStorage(t *testing.T) {
	//
	collName := fmt.Sprintf("feeds-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "producer-rss",
	}
	dbCfg.Table.Name = collName
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	assert.NotNil(t, s)
	assert.Nil(t, err)
	//
	clear(ctx, t, s.(storageMongo))
}

func clear(ctx context.Context, t *testing.T, sm storageMongo) {
	require.Nil(t, sm.coll.Drop(ctx))
	require.Nil(t, sm.Close())
}

func TestStorageMongo_GetUpdateTime(t *testing.T) {
	//
	collName := fmt.Sprintf("feeds-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "producer-rss",
	}
	dbCfg.Table.Name = collName
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.NotNil(t, s)
	require.Nil(t, err)
	sm := s.(storageMongo)
	defer clear(ctx, t, sm)
	//
	_, err = sm.coll.InsertOne(ctx, feedRec{
		Url:        "https://test.rss.com",
		UpdateTime: time.Date(2023, 5, 23, 8, 52, 40, 0, time.UTC),
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		url string
		ut  time.Time
	}{
		"found": {
			url: "https://test.rss.com",
			ut:  time.Date(2023, 5, 23, 8, 52, 40, 0, time.UTC),
		},
		"not found": {
			url: "https://missing.com",
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var ut time.Time
			ut, err = sm.GetUpdateTime(ctx, c.url)
			assert.Equal(t, c.ut, ut)
			assert.Nil(t, err)
		})
	}
}

func TestStorageMongo_SetUpdateTime(t *testing.T) {
	//
	collName := fmt.Sprintf("feeds-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "producer-rss",
	}
	dbCfg.Table.Name = collName
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.NotNil(t, s)
	require.Nil(t, err)
	sm := s.(storageMongo)
	defer clear(ctx, t, sm)
	//
	_, err = sm.coll.InsertOne(ctx, feedRec{
		Url:        "https://test0.rss.com",
		UpdateTime: time.Date(2023, 5, 23, 8, 52, 40, 0, time.UTC),
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		url string
		ut  time.Time
	}{
		"existing": {
			url: "https://test0.rss.com",
			ut:  time.Date(2023, 5, 23, 9, 00, 40, 0, time.UTC),
		},
		"new": {
			url: "https://test1.rss.com",
			ut:  time.Date(2023, 5, 23, 9, 00, 40, 0, time.UTC),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err = sm.SetUpdateTime(ctx, c.url, c.ut)
			assert.Nil(t, err)
		})
	}
}
