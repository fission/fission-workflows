// Note: package is called 'native' because 'internal' is not an allowed package name.
package native

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/ptypes"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const (
	Name = "native"
)

var (
	fnActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "fnenv",
		Subsystem: "native",
		Name:      "functions_active",
		Help:      "Number of function executions that are currently active",
	})

	fnCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "fnenv",
		Subsystem: "native",
		Name:      "functions_execution_total",
		Help:      "Total number of function executions",
	})

	fnResolved = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "fnenv",
		Subsystem: "native",
		Name:      "functions_resolved_total",
		Help:      "Total number of function resolved",
	})

	fnExecTime = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "fnenv",
		Subsystem: "native",
		Name:      "function_execution_time_milliseconds",
		Help:      "Execution time summary of the internal functions",
	})
)

func init() {
	prometheus.MustRegister(fnActive, fnResolved, fnExecTime, fnCount)
}

// An InternalFunction is a function that will be executed in the same process as the invoker.
type InternalFunction interface {
	Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error)
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
	log.WithField("fns", env.fns).Debugf("Internal function runtime installed.")
	return env
}

func (fe *FunctionEnv) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error) {
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
	defer fnExecTime.Observe(float64(time.Since(timeStart)))
	fnID := spec.FnRef.ID
	fn, ok := fe.fns[fnID]
	if !ok {
		return nil, fmt.Errorf("could not resolve internal function '%s'", fnID)
	}
	fnActive.Inc()
	out, err := fn.Invoke(spec)
	fnActive.Dec()
	fnCount.Inc()
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

func (fe *FunctionEnv) Resolve(fnName string) (string, error) {
	fnResolved.Inc()
	_, ok := fe.fns[fnName]
	if !ok {
		return "", fmt.Errorf("could not resolve internal function '%s'", fnName)
	}
	log.WithField("name", fnName).WithField("uid", fnName).Debug("Resolved internal function")
	return fnName, nil
}

func (fe *FunctionEnv) RegisterFn(name string, fn InternalFunction) {
	fe.fns[name] = fn
}
