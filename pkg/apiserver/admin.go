package apiserver

import (
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/version"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

type Admin struct {
	metaCtrl controller.MetaController
}

func (as *Admin) Status(ctx context.Context, _ *empty.Empty) (*Health, error) {
	return &Health{
		Status: "OK!",
	}, nil
}

func (as *Admin) Version(ctx context.Context, _ *empty.Empty) (*version.Info, error) {
	v := version.VersionInfo()
	return &v, nil
}
