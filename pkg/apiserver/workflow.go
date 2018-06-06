package apiserver

import (
	"errors"
	"strings"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"

	"golang.org/x/net/context"
)

type Workflow struct {
	api   *api.Workflow
	cache fes.CacheReader
}

func NewWorkflow(api *api.Workflow, cache fes.CacheReader) *Workflow {
	wf := &Workflow{
		api:   api,
		cache: cache,
	}

	return wf
}

func (ga *Workflow) Create(ctx context.Context, spec *types.WorkflowSpec) (*WorkflowIdentifier, error) {

	id, err := ga.api.Create(spec)
	if err != nil {
		logrus.Info(strings.Replace(validate.Format(err), "\n", "; ", -1))
		return nil, err
	}

	return &WorkflowIdentifier{id}, nil
}

func (ga *Workflow) Get(ctx context.Context, workflowID *WorkflowIdentifier) (*types.Workflow, error) {
	id := workflowID.GetId()
	if len(id) == 0 {
		return nil, errors.New("no id provided")
	}

	entity := aggregates.NewWorkflow(id)
	err := ga.cache.Get(entity)
	if err != nil {
		return nil, err
	}
	return entity.Workflow, nil
}

func (ga *Workflow) Delete(ctx context.Context, workflowID *WorkflowIdentifier) (*empty.Empty, error) {
	id := workflowID.GetId()
	if len(id) == 0 {
		return nil, errors.New("no id provided")
	}

	err := ga.api.Delete(id)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (ga *Workflow) List(ctx context.Context, req *empty.Empty) (*SearchWorkflowResponse, error) {
	var results []string
	wfs := ga.cache.List()
	for _, result := range wfs {
		results = append(results, result.Id)
	}
	return &SearchWorkflowResponse{results}, nil
}

func (ga *Workflow) Validate(ctx context.Context, spec *types.WorkflowSpec) (*empty.Empty, error) {
	err := validate.WorkflowSpec(spec)
	if err != nil {
		logrus.Info(strings.Replace(validate.Format(err), "\n", "; ", -1))
		return nil, err
	}
	return &empty.Empty{}, nil
}
