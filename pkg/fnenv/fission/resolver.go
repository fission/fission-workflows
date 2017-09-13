package fission

import (
	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission/controller/client"
	"k8s.io/client-go/1.5/pkg/api"
)

type Resolver struct {
	controller *client.Client
}

func NewResolver(controller *client.Client) *Resolver {
	return &Resolver{controller}
}

func (re *Resolver) Resolve(fnName string) (string, error) {
	fn, err := re.controller.FunctionGet(&api.ObjectMeta{
		Name: fnName,
	})
	if err != nil {
		return "", err
	}

	return string(fn.Metadata.UID), nil
}
