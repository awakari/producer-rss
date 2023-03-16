package service

import "net/http"

type Client interface {
	Get(url string) (resp *http.Response, err error)
}

type client struct {
	httpClient http.Client
	userAgent  string
}

func NewClient(httpClient http.Client, userAgent string) Client {
	return client{
		httpClient: httpClient,
		userAgent:  userAgent,
	}
}

func (c client) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	return c.httpClient.Do(req)
}
