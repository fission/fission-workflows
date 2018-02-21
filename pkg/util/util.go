package util

import (
	"fmt"
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

// Uid generates a unique id
func Uid() string {
	return uuid.NewV4().String()
}

func CreateScopeId(parentId string, taskId string) string {
	if len(taskId) == 0 {
		taskId = Uid()
	}
	return fmt.Sprintf("%s_%s", parentId, taskId)
}

// Wait is a sync.WaitGroup.Wait() implementation that supports timeouts
func Wait(wg *sync.WaitGroup, timeout time.Duration) bool {
	wgDone := make(chan bool)
	defer close(wgDone)
	go func() {
		wg.Wait()
		wgDone <- true
	}()

	select {
	case <-wgDone:
		return true
	case <-time.After(timeout):
		return false
	}
}
