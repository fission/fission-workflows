package controller

import (
	"time"

	"github.com/sirupsen/logrus"
)

type evalContext struct {
	state *EvalState
}

func NewEvalContext(state *EvalState) EvalContext {
	return &evalContext{
		state: state,
	}
}

func (ec *evalContext) EvalState() *EvalState {
	return ec.state
}

type RuleTimedOut struct {
	OnTimedOut   Rule
	OnWithinTime Rule
	Timeout      time.Duration
}

func (tf *RuleTimedOut) Eval(ec EvalContext) Action {
	status := ec.EvalState()
	initialStatus, ok := status.First()
	if !ok {
		// Invocation has not yet started
		return evalIfNotNil(tf.OnWithinTime, ec)
	}
	duration := time.Now().UnixNano() - initialStatus.Timestamp.UnixNano()
	if duration > tf.Timeout.Nanoseconds() {
		return evalIfNotNil(tf.OnTimedOut, ec)
	}
	return evalIfNotNil(tf.OnWithinTime, ec)
}

type RuleExceededErrorCount struct {
	OnExceeded    Rule
	OnNotExceeded Rule
	MaxErrorCount int
}

func (el *RuleExceededErrorCount) Eval(ec EvalContext) Action {
	var errorCount int
	state := ec.EvalState()
	for i := state.Count() - 1; i >= 0; i-- {
		record, ok := state.Get(i)
		if !ok {
			panic("Illegal modification")
		}
		if record.Error == nil {
			break
		}
		errorCount++
	}

	if errorCount > el.MaxErrorCount {
		logrus.Infof("Error count exceeded, evaluating %T", el.OnExceeded)
		return evalIfNotNil(el.OnExceeded, ec)
	}
	return evalIfNotNil(el.OnNotExceeded, ec)
}

type RuleEvalUntilAction struct {
	Rules []Rule
}

func (cf *RuleEvalUntilAction) Eval(ec EvalContext) Action {
	for _, i := range cf.Rules {
		if i == nil {
			continue
		}
		action := i.Eval(ec)
		if action != nil {
			return action
		}
	}
	return nil
}

func evalIfNotNil(rule Rule, ec EvalContext) Action {
	if rule == nil {
		return nil
	}
	return rule.Eval(ec)
}
