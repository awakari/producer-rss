package resolver

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/v2/event"
	"golang.org/x/exp/slog"
)

type (
	loggingMiddleware struct {
		svc Service
		log *slog.Logger
	}
)

func NewLoggingMiddleware(svc Service, log *slog.Logger) Service {
	return loggingMiddleware{
		svc: svc,
		log: log,
	}
}

func (lm loggingMiddleware) SubmitBatch(ctx context.Context, msgs []*event.Event) (count uint32, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("resolver.SubmitBatch(count=%d): %d, %s", len(msgs), count, err))
	}()
	return lm.svc.SubmitBatch(ctx, msgs)
}
