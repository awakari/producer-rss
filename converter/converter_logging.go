package converter

import (
	"fmt"
	"github.com/SlyMarbo/rss"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"golang.org/x/exp/slog"
)

type converterLogging struct {
	conv Converter
	log  *slog.Logger
}

func NewConverterLogging(conv Converter, log *slog.Logger) Converter {
	return converterLogging{
		conv: conv,
		log:  log,
	}
}

func (cl converterLogging) Convert(feed *rss.Feed, item *rss.Item) (msg *pb.CloudEvent) {
	msg = cl.conv.Convert(feed, item)
	cl.log.Debug(fmt.Sprintf("converter.Convert(_, %s): %s", item.ID, msg.Id))
	return
}
