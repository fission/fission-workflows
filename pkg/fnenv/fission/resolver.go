package fission

import (
	"github.com/fission/fission"
	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission/controller/client"
)

type Registry struct {
	client *client.Client
}

func NewResolver(client *client.Client) function.Resolver {
	return &Registry{client}
}

func (re *Registry) Resolve(fnName string) (string, error) {
	fn, err := re.client.FunctionGet(&fission.Metadata{
		Name: fnName,
	})
	if err != nil {
		return "", err
	}

	return fn.Uid, nil
}
