package apiserver

import (
	"github.com/fission/fission-workflow/pkg/types"
	"golang.org/x/net/context"
)

type GrpcAdminApiServer struct {
}

func (as *GrpcAdminApiServer) Status(ctx context.Context, _ *types.Empty) (*Health, error) {

	return &Health{
		Status: "OK!",
	}, nil
}
