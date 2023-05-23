package main

import (
	"context"
	"fmt"
	"golang.org/x/exp/slog"
	"time"
)

type producerLogging struct {
	prod Producer
	log  *slog.Logger
}

func NewProducerLogging(prod Producer, log *slog.Logger) Producer {
	return producerLogging{
		prod: prod,
		log:  log,
	}
}

func (pl producerLogging) Produce(ctx context.Context) (nextTime time.Time, err error) {
	nextTime, err = pl.prod.Produce(ctx)
	if err == nil {
		pl.log.Debug(fmt.Sprintf("producer.Produce(_): %s", nextTime.Format(time.RFC3339)))
	} else {
		pl.log.Warn(fmt.Sprintf("producer.Produce(_): %s, %s", nextTime.Format(time.RFC3339), err))
	}
	return
}
