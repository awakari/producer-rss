package feeds

import (
	"context"
	"errors"
	"io"
	"time"
)

type Storage interface {
	io.Closer
	GetUpdateTime(ctx context.Context, url string) (t time.Time, err error)
	SetUpdateTime(ctx context.Context, url string, t time.Time) (err error)
}

var ErrInternal = errors.New("internal failure")
