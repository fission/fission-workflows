package actions

import (
	"github.com/fission/fission-workflows/pkg/api/invocation"
	"github.com/sirupsen/logrus"
)

func Abort(invocationId string, api *invocation.Api) error {
	logrus.Infof("aborting: '%v'", invocationId)
	return api.Cancel(invocationId)
}
