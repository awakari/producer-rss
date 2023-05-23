package feeds

import (
	"fmt"
	"golang.org/x/exp/slog"
	"net/http"
)

type loggingMiddleware struct {
	client Client
	log    *slog.Logger
}

func NewLoggingMiddleware(client Client, log *slog.Logger) Client {
	return loggingMiddleware{
		client: client,
		log:    log,
	}
}

func (lm loggingMiddleware) Get(url string) (resp *http.Response, err error) {
	defer func() {
		var statusCode int
		if resp != nil {
			statusCode = resp.StatusCode
		}
		lm.log.Debug(fmt.Sprintf("client.Get(url=%s): %d, %s", url, statusCode, err))
	}()
	return lm.client.Get(url)
}
