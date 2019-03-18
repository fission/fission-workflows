package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLocalExecutor(t *testing.T) {
	executor := NewLocalExecutor(1, 3)
	t1 := &testTask{0}
	t2 := &testTask{1}
	t3 := &testTask{2}
	t4 := &testTask{3}
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
	assert.Equal(t, 1, t1.n)
	assert.Equal(t, 2, t2.n)
	assert.Equal(t, 3, t3.n)
}

type testTask struct {
	n int
}

func (t *testTask) Apply() error {
	t.n++
	return nil
}
