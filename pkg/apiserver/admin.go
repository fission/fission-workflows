package apiserver

import (
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/version"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

type GrpcAdminApiServer struct {
	metaCtrl controller.MetaController
}

func (as *GrpcAdminApiServer) Status(ctx context.Context, _ *empty.Empty) (*Health, error) {
	return &Health{
		Status: "OK!",
	}, nil
}

func (as *GrpcAdminApiServer) Version(ctx context.Context, _ *empty.Empty) (*version.Info, error) {
	version := version.VersionInfo()
	return &version, nil
}

func (as *GrpcAdminApiServer) Resume(context.Context, *empty.Empty) (*empty.Empty, error) {
	as.metaCtrl.Resume()
	return &empty.Empty{}, nil
}

func (as *GrpcAdminApiServer) Halt(context.Context, *empty.Empty) (*empty.Empty, error) {
	as.metaCtrl.Halt()
	return &empty.Empty{}, nil
}
