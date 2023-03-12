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
		lm.log.Debug(
			fmt.Sprintf(
				"resolver.Submit(msg(id: \"%s\", source: \"%s\" subject: \"%s\" time: %s)): %s",
				msg.ID(),
				msg.Source(),
				msg.Subject(),
				msg.Time(),
				err,
			),
		)
	}()
	return lm.svc.Submit(ctx, msg)
}
