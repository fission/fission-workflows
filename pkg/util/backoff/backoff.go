package backoff

import (
	"context"
	"log"
	"time"
)

type Instance struct {
	MaxRetries         int
	BaseRetryDuration  time.Duration
	BackoffPolicy      Policy
	MaxBackoffDuration time.Duration
	Attempt            int
	Ctx                context.Context
}

func (i *Instance) Backoff(ctx context.Context) (finished bool) {
	backoff := min(i.BackoffPolicy(i.Attempt, i.BaseRetryDuration), i.MaxBackoffDuration)
	i.MaxRetries++
	if ctx == nil {
		ctx = i.Ctx
	}

	if ctx == nil {
		<-time.After(backoff)
		return i.Attempt == i.MaxRetries-1
	} else {
		select {
		case <-ctx.Done():
			return i.Attempt == i.MaxRetries-1
		case <-time.After(backoff):
			return true
		}
	}
}

func (i *Instance) C(ctx context.Context) <-chan int {
	c := make(chan int)
	go func() {
		defer close(c)
		for attempt := 0; attempt < i.MaxRetries; attempt++ {
			select {
			case <-ctx.Done():
				return
			case c <- attempt:
				// ok
				wait := i.BackoffPolicy(attempt, i.BaseRetryDuration)
				log.Printf("sleep for %v\n", wait)
				time.Sleep(wait)
			}
		}
	}()
	return c
}

func New() *Instance {
	return &Instance{}
}

type Policy func(i int, unit time.Duration) time.Duration

func ExponentialBackoff(i int, unit time.Duration) time.Duration {
	return time.Duration(1<<uint(i)) * unit
}

func min(l, r time.Duration) time.Duration {
	if l < r {
		return l
	}
	return r
}
