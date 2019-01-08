package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

// SyncMapLen is simply a sync.Map with options to retrieve the length of it.
type SyncMapLen struct {
	mp        sync.Map
	cachedLen *int32
}

func (e *SyncMapLen) Len() int {
	if e.cachedLen == nil {
		var count int32
		e.mp.Range(func(k, v interface{}) bool {
			count++
			return true
		})
		atomic.StoreInt32(e.cachedLen, count)
		return int(count)
	}
	return int(atomic.LoadInt32(e.cachedLen))
}

func (e *SyncMapLen) LoadOrStore(key interface{}, value interface{}) (actual interface{}, loaded bool) {
	actual, loaded = e.mp.LoadOrStore(key, value)
	if !loaded && e.cachedLen != nil {
		atomic.AddInt32(e.cachedLen, 1)
	}
	return actual, loaded
}

func (e *SyncMapLen) Load(key interface{}) (value interface{}, ok bool) {
	return e.mp.Load(key)
}

func (e *SyncMapLen) Store(key, value interface{}) {
	e.mp.Store(key, value)
	e.cachedLen = nil // We are not sure if an entry was replaced or added
}

func (e *SyncMapLen) Delete(id string) {
	e.mp.Delete(id)
	e.cachedLen = nil // We are not sure if an entry was removed
}

func (e *SyncMapLen) Range(f func(key interface{}, value interface{}) bool) {
	e.mp.Range(f)
}

func LogIfError(err error) {
	if err != nil {
		logrus.Error(err)
	}
}

func AssertProtoEqual(t *testing.T, expected, actual proto.Message) {
	assert.True(t, proto.Equal(expected, actual), "expected: %v, actual: %v", expected, actual)
}

// Numeric is a representation
type Number struct {
	val float64 // Fix loss of precision in uint64 and int64
}

func (n Number) Value() interface{} {
	// TODO return original type
	return n.val
}

func ToNumber(val interface{}) (Number, error) {
	switch t := val.(type) {
	case float64:
		return Number{val: t}, nil
	case float32:
		return Number{val: float64(t)}, nil
	case int:
		return Number{val: float64(t)}, nil
	case int32:
		return Number{val: float64(t)}, nil
	case int16:
		return Number{val: float64(t)}, nil
	case int64:
		return Number{val: float64(t)}, nil
	case int8:
		return Number{val: float64(t)}, nil
	default:
		return Number{}, errors.New("not a supported number (int, int8, int16, int32, int64, float32, " +
			"and float64)")
	}
}

func CmpProtoTimestamps(l, r *timestamp.Timestamp) bool {
	if l.GetSeconds() < r.GetSeconds() {
		return true
	} else if l.GetSeconds() > r.GetSeconds() {
		return false
	}
	return l.GetNanos() <= r.GetNanos()
}
