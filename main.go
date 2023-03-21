package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"os"
	"producer-rss/api/grpc/resolver"
	"producer-rss/config"
	"producer-rss/service"
)

func main() {
	//
	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("failed to load the config from env: %s", err))
	}
	opts := slog.HandlerOptions{
		Level: slog.Level(cfg.Log.Level),
	}
	log := slog.New(opts.NewTextHandler(os.Stdout))
	log.Info("starting...")
	//
	if len(os.Args) != 2 {
		panic("invalid command line arguments, try: producer-rss <feeds-file-path>")
	}
	cfgFilePath := os.Args[1]
	slog.Info(fmt.Sprintf("loading the feeds from the file: %s ...", cfgFilePath))
	cfgFile, err := os.Open(cfgFilePath)
	if err != nil {
		panic(fmt.Sprintf("failed to open the feeds file %s: %s", cfgFilePath, err))
	}
	s := bufio.NewScanner(cfgFile)
	for s.Scan() {
		feed := s.Text()
		slog.Info("loaded feed URL: " + feed)
		cfg.Feed.Urls = append(cfg.Feed.Urls, feed)
	}
	//
	resolverConn, err := grpc.Dial(
		cfg.Api.Resolver.Uri,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	resolverClient := resolver.NewServiceClient(resolverConn)
	resolverSvc := resolver.NewService(resolverClient)
	// resolverSvc := resolver.NewServiceMock() // uncomment for the local testing
	resolverSvc = resolver.NewLoggingMiddleware(resolverSvc, log)
	//
	httpClient := http.Client{
		Timeout: cfg.Feed.UpdateTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.Feed.TlsSkipVerify,
			},
		},
	}
	feedsClient := service.NewClient(httpClient, cfg.Feed.UserAgent)
	feedsClient = service.NewLoggingMiddleware(feedsClient, log)
	//
	svc, err := service.NewService(cfg.Feed, feedsClient, cfg.Message, cfg.Api.Resolver.Backoff, resolverSvc)
	if err != nil {
		log.Error("service init error", err)
	}
	errChan := make(chan error)
	log.Info("starting the feeds processing")
	go svc.ProcessLoop(errChan)
	for {
		err = <-errChan
		log.Error("processing failures", err)
	}
}
