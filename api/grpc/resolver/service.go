package resolver

import (
	"context"
	"errors"
	"fmt"
	format "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/cloudevents/sdk-go/v2/event"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	Submit(ctx context.Context, msg *event.Event) (err error)
}

type service struct {
	client ServiceClient
}

// ErrInternal indicates some unexpected internal failure.
var ErrInternal = errors.New("internal failure")

var ErrQueueMissing = errors.New("missing queue")

var ErrQueueFull = errors.New("queue is full")

func NewService(client ServiceClient) Service {
	return service{
		client: client,
	}
}

func (svc service) Submit(ctx context.Context, msg *event.Event) (err error) {
	var msgProto *pb.CloudEvent
	msgProto, err = format.ToProto(msg)
	if err == nil {
		_, err = svc.client.Submit(ctx, msgProto)
		if err != nil {
			err = decodeError(err)
		}
	}
	return
}

func decodeError(src error) (dst error) {
	switch status.Code(src) {
	case codes.OK:
		dst = nil
	case codes.NotFound:
		dst = fmt.Errorf("%w: resolver: %s", ErrQueueMissing, src)
	case codes.ResourceExhausted:
		dst = fmt.Errorf("%w: resolver: %s", ErrQueueFull, src)
	default:
		dst = fmt.Errorf("%w: resolver: %s", ErrInternal, src)
	}
	return
}
