package util

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

// UID generates a unique id
func UID() string {
	return uuid.NewV4().String()
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

func Truncate(val interface{}, maxLen int) string {
	s := fmt.Sprintf("%v", val)
	if len(s) <= maxLen {
		return s
	}
	affix := fmt.Sprintf("<truncated orig_len: %d>", len(s))
	return s[:len(s)-len(affix)] + affix
}
