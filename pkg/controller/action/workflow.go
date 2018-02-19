package action

import (
	"github.com/fission/fission-workflows/pkg/api/workflow"
	"github.com/fission/fission-workflows/pkg/types"
)

type ParseWorkflow struct {
	WfApi *workflow.Api
	Wf    *types.Workflow
}

func (ac *ParseWorkflow) Id() string {
	return ac.Wf.Metadata.Id
}

func (ac *ParseWorkflow) Apply() error {
	_, err := ac.WfApi.Parse(ac.Wf)
	return err
}
