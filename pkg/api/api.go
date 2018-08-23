package api

import (
	"context"
	"errors"

	"github.com/fission/fission-workflows/pkg/types"
)

type CallConfig struct {
	ctx             context.Context
	postTransformer func(i interface{}) error
}

type CallOption func(op *CallConfig)

func WithContext(ctx context.Context) CallOption {
	return func(op *CallConfig) {
		op.ctx = ctx
	}
}

func PostTransformer(fn func(ti *types.TaskInvocation) error) CallOption {
	return func(op *CallConfig) {
		op.postTransformer = func(i interface{}) error {
			ti, ok := i.(*types.TaskInvocation)
			if !ok {
				return errors.New("invalid call option")
			}
			return fn(ti)
		}
	}
}

func parseCallOptions(opts []CallOption) *CallConfig {
	// Default
	cfg := &CallConfig{
		ctx: context.Background(),
	}
	// Parse options
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
