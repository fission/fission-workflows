package fission

import (
	"github.com/fission/fission/controller/client"
	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resolver struct {
	controller *client.Client
}

func NewResolver(controller *client.Client) *Resolver {
	return &Resolver{controller}
}

func (re *Resolver) Resolve(fnName string) (string, error) {
	logrus.WithField("name", fnName).Info("Resolving function ")
	fn, err := re.controller.FunctionGet(&metav1.ObjectMeta{
		Name: fnName,
	})
	if err != nil {
		return "", err
	}
	id := string(fn.Metadata.UID)

	logrus.WithField("name", fnName).WithField("uid", id).Info("Resolved fission function")

	return id, nil
}
