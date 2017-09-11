package fission

import (
	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission/controller/client"
	"k8s.io/client-go/1.5/pkg/api"
)

type Registry struct {
	client *client.Client
}

func NewResolver(client *client.Client) function.Resolver {
	return &Registry{client}
}

func (re *Registry) Resolve(fnName string) (string, error) {
	fn, err := re.client.FunctionGet(&api.ObjectMeta{
		Name: fnName,
	})
	if err != nil {
		return "", err
	}

	return string(fn.Metadata.UID), nil
}
