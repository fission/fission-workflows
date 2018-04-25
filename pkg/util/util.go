package util

import (
	"encoding/json"
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

func ConvertStructsToMap(i interface{}) (map[string]interface{}, error) {
	ds, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	var mp map[string]interface{}
	err = json.Unmarshal(ds, &mp)
	if err != nil {
		return nil, err
	}
	return mp, nil
}

func MustConvertStructsToMap(i interface{}) map[string]interface{} {
	if result, err := ConvertStructsToMap(i); err != nil {
		panic(err)
	} else {
		return result
	}
}
