/*
Package apiserver contains all request handlers for gRPC and HTTP servers.

It has these top-level request handlers:
	Admin	   - Administrative functionality related to managing the workflow engine.
	Invocation - Functionality related to managing invocations.
	Workflow   - functionality related to managing workflows.

The purpose of this package is purely to provide handlers to gRPC and HTTP servers. Therefore,
it should not contain any logic (validation, composition, etc.) related to the workflows,
invocations or any other targets that it provides. All this logic should be placed in the actual
packages that are responsible for the business logic, such as `api`.
*/
package apiserver

import (
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Empty = empty.Empty

func toErrorStatus(err error) error {
	switch err.(type) {
	case validate.Error:
		logrus.Errorf("Request error: %v", validate.FormatConcise(err))
		return status.Error(codes.InvalidArgument, validate.Format(err))
	default:
		logrus.Errorf("Request error: %v", err)
		return err
	}
}
