package fission

import (
	"github.com/fission/fission/controller/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resolver implements the Resolver interface and is responsible for resolving function references to
// deterministic Fission function UIDs.
type Resolver struct {
	controller *client.Client
}

func NewResolver(controller *client.Client) *Resolver {
	return &Resolver{controller}
}

func (re *Resolver) Resolve(fnName string) (string, error) {
	log.Infof("Resolving function: %s", fnName)
	fn, err := re.controller.FunctionGet(&metav1.ObjectMeta{
		Name:      fnName,
		Namespace: metav1.NamespaceDefault,
	})
	if err != nil {
		return "", err
	}
	id := fnName + "-" + string(fn.Metadata.UID)

	log.Infof("Resolved fission function %s to %s", fnName, id)

	return id, nil
}
