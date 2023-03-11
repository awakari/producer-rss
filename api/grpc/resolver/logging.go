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

func (lm loggingMiddleware) Submit(ctx context.Context, msg *event.Event) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("resolver.Submit(msg.Id=%s): %s", msg.ID(), err))
	}()
	return lm.svc.Submit(ctx, msg)
}
