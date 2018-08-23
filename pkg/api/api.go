package api

import "context"

type CallConfig struct {
	ctx context.Context
}

type CallOption func(op *CallConfig)

func WithContext(ctx context.Context) CallOption {
	return func(op *CallConfig) {
		op.ctx = ctx
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
