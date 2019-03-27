package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
)

func TestLocalExecutor(t *testing.T) {
	executor := NewLocalExecutor(1, 3)

	t1 := &testTask{atomic.NewInt32(0)}
	t2 := &testTask{atomic.NewInt32(1)}
	t3 := &testTask{atomic.NewInt32(2)}
	t4 := &testTask{atomic.NewInt32(3)}
	accepted := executor.Submit(&Task{
		Apply: t1.Apply,
	})
	assert.True(t, accepted)
	accepted = executor.Submit(&Task{
		Apply: t2.Apply,
	})
	assert.True(t, accepted)
	accepted = executor.Submit(&Task{
		Apply: t3.Apply,
	})
	assert.True(t, accepted)
	accepted = executor.Submit(&Task{
		Apply: t4.Apply,
	})
	assert.False(t, accepted)
	assert.Equal(t, 3, executor.queue.Len())
	executor.Start()
	defer executor.Close()
	time.Sleep(time.Second) // wait to complete
	assert.Equal(t, int32(1), t1.n.Load())
	assert.Equal(t, int32(2), t2.n.Load())
	assert.Equal(t, int32(3), t3.n.Load())
}

type testTask struct {
	n *atomic.Int32
}

func (t *testTask) Apply() error {
	t.n.Add(1)
	return nil
}
