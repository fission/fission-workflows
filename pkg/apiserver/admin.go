package apiserver

import (
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

type GrpcAdminApiServer struct {
}

func (as *GrpcAdminApiServer) Status(ctx context.Context, _ *empty.Empty) (*Health, error) {

	return &Health{
		Status: "OK!",
	}, nil
}
