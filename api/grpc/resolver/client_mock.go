package resolver

import (
	"context"
	"google.golang.org/grpc"
)

type clientMock struct {
}

func NewClientMock() ServiceClient {
	return clientMock{}
}

func (cm clientMock) SubmitBatch(ctx context.Context, in *SubmitBatchRequest, opts ...grpc.CallOption) (resp *BatchResponse, err error) {
	resp = &BatchResponse{}
	for _, msg := range in.Msgs {
		if msg.Id == "fail" {
			resp.Err = ErrInternal.Error()
			break
		}
		if msg.Id == "missing" {
			resp.Err = ErrQueueMissing.Error()
			break
		}
		if msg.Id == "full" {
			break
		}
		resp.Count++
	}
	return resp, nil
}
