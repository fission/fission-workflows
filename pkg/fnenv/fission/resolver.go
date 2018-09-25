package fission

import (
	"github.com/fission/fission-workflows/pkg/types"
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

func (re *Resolver) Resolve(ref types.FnRef) (string, error) {
	// Currently we just use the controller API to check if the function exists.
	log.Infof("Resolving function: %s", ref.ID)
	ns := ref.Namespace
	if len(ns) == 0 {
		ns = metav1.NamespaceDefault
	}
	_, err := re.controller.FunctionGet(&metav1.ObjectMeta{
		Name:      ref.ID,
		Namespace: ns,
	})
	if err != nil {
		return "", err
	}
	id := ref.ID

	log.Infof("Resolved fission function %s to %s", ref.ID, id)
	return id, nil
}
