package controller

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockRule struct {
	count *int
}

func (r *MockRule) Eval(cec EvalContext) []Action {
	*r.count++
	return nil
}

func TestRuleExceededErrorCount_Eval(t *testing.T) {
	exceeded := 0
	notExceeded := 0
	rule := RuleExceededErrorCount{
		OnExceeded:    &MockRule{&exceeded},
		OnNotExceeded: &MockRule{&notExceeded},
		MaxErrorCount: 0,
	}

	// Not exceeded
	es := NewEvalState("randomId", nil)
	ctx := NewEvalContext(es)
	rule.Eval(ctx)
	assert.Equal(t, exceeded, 0)
	assert.Equal(t, notExceeded, 1)

	es.Record(EvalRecord{
		Cause: "t",
		Error: errors.New("some error"),
	})
	rule.Eval(ctx)
	assert.Equal(t, exceeded, 1)
	assert.Equal(t, notExceeded, 1)
}
