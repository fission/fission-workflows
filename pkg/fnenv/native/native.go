// Note: package is called 'native' because 'internal' is not an allowed package name.
package native

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/ptypes"
	"github.com/opentracing/opentracing-go"

	log "github.com/sirupsen/logrus"
)

const (
	Name = "native"
)

// An InternalFunction is a function that will be executed in the same process as the invoker.
type InternalFunction interface {
	Invoke(spec *types.TaskInvocationSpec) (*typedvalues.TypedValue, error)
}

// FunctionEnv for executing low overhead functions, such as control flow constructs, inside the workflow engine
//
// Note: This currently supports Golang only.
type FunctionEnv struct {
	fns map[string]InternalFunction // Name -> function
}

func NewFunctionEnv(fns map[string]InternalFunction) *FunctionEnv {
	env := &FunctionEnv{
		fns: fns,
	}
	return env
}

func (fe *FunctionEnv) Invoke(spec *types.TaskInvocationSpec, opts ...fnenv.InvokeOption) (*types.TaskInvocationStatus, error) {
	cfg := fnenv.ParseInvokeOptions(opts)
	defer func() {
		if r := recover(); r != nil {
			log.WithFields(log.Fields{
				"err": r,
			}).Error("Internal function crashed.")
			fmt.Println(string(debug.Stack()))
		}
	}()
	if err := validate.TaskInvocationSpec(spec); err != nil {
		return nil, err
	}

	timeStart := time.Now()
	defer fnenv.FnExecTime.WithLabelValues(Name).Observe(float64(time.Since(timeStart)))
	fnID := spec.FnRef.ID
	fn, ok := fe.fns[fnID]
	if !ok {
		return nil, fmt.Errorf("could not resolve internal function '%s'", fnID)
	}
	span, _ := opentracing.StartSpanFromContext(cfg.Ctx, fmt.Sprintf("fnenv/internal/fn/%s", fnID))
	defer span.Finish()
	fnenv.FnActive.WithLabelValues(Name).Inc()
	out, err := fn.Invoke(spec)
	fnenv.FnActive.WithLabelValues(Name).Dec()
	fnenv.FnCount.WithLabelValues(Name).Inc()
	if err != nil {
		log.WithFields(log.Fields{
			"fnID": fnID,
			"err":  err,
		}).Error("Internal function failed.")
		return &types.TaskInvocationStatus{
			UpdatedAt: ptypes.TimestampNow(),
			Status:    types.TaskInvocationStatus_FAILED,
			Error: &types.Error{
				Message: err.Error(),
			},
		}, nil
	}

	return &types.TaskInvocationStatus{
		UpdatedAt: ptypes.TimestampNow(),
		Status:    types.TaskInvocationStatus_SUCCEEDED,
		Output:    out,
	}, nil
}

func (fe *FunctionEnv) Resolve(ref types.FnRef) (string, error) {
	_, ok := fe.fns[ref.ID]
	if !ok {
		return "", fmt.Errorf("could not resolve internal function '%s'", ref.ID)
	}
	log.WithField("name", ref.ID).WithField("uid", ref.ID).Debug("Resolved internal function")
	return ref.ID, nil
}

func (fe *FunctionEnv) RegisterFn(name string, fn InternalFunction) {
	fe.fns[name] = fn
}

// Installed lists all installed functions in the internal function runtime.
func (fe *FunctionEnv) Installed() []string {
	fns := make([]string, len(fe.fns))
	for fn := range fe.fns {
		fns = append(fns, fn)
	}
	return fns
}
