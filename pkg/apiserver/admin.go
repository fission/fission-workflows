package apiserver

import (
	"github.com/fission/fission-workflows/pkg/version"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

const StatusOK = "OK!"

// Admin is responsible for all administrative functions related to managing the workflow engine.
type Admin struct {
}

func (as *Admin) Status(ctx context.Context, _ *empty.Empty) (*Health, error) {
	return &Health{
		Status: StatusOK,
	}, nil
}

func (as *Admin) Version(ctx context.Context, _ *empty.Empty) (*version.Info, error) {
	v := version.VersionInfo()
	return &v, nil
}
