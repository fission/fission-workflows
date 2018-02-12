package apiserver

import (
	"github.com/fission/fission-workflows/pkg/version"
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

func (as *GrpcAdminApiServer) Version(ctx context.Context, _ *empty.Empty) (*VersionResp, error) {

	return &VersionResp{
		Version: version.VERSION,
	}, nil
}
