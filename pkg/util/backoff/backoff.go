package backoff

import (
	"math"
	"time"
)

type Algorithm interface {
	Backoff(context ...Context) Context
}

var DefaultBackoffAlgorithm = &ExponentialBackoff{
	MaxBackoff: time.Duration(24) * time.Hour,
	Exponent:   2,
	Step:       time.Duration(100) * time.Millisecond,
}

func Backoff(context ...Context) Context {
	return DefaultBackoffAlgorithm.Backoff(context...)
}

type Context struct {
	Lockout   time.Time
	Attempts  int
	Algorithm Algorithm
}

func (bc Context) Next() Context {
	// For now just exponential
	alg := bc.Algorithm
	if alg == nil {
		alg = DefaultBackoffAlgorithm
	}
	return alg.Backoff(bc)
}

func (bc Context) Locked(time time.Time) bool {
	return bc.Lockout.After(time)
}

type ExponentialBackoff struct {
	MaxBackoff time.Duration
	MinBackoff time.Duration
	Step       time.Duration
	Exponent   int
}

func (eb *ExponentialBackoff) Backoff(context ...Context) Context {
	// Base case
	c := 0
	l := time.Now()
	if len(context) != 0 {
		c = context[0].Attempts
		l = context[0].Lockout
	}

	max := eb.MaxBackoff.Nanoseconds()
	min := eb.MinBackoff.Nanoseconds()
	step := eb.MinBackoff.Nanoseconds()
	if max == 0 {
		max = math.MaxInt64
	}

	d := (0.5 * (math.Pow(float64(2), float64(c)) - 1.0)) * float64(step)

	lockoutDuration := math.Max(math.Min(d, float64(max)), float64(min))

	return Context{
		Algorithm: eb,
		Attempts:  c + 1,
		Lockout:   l.Add(time.Duration(lockoutDuration) * time.Nanosecond),
	}
}

type Map map[string]Context

func NewMap() Map {
	return Map(map[string]Context{})
}

func (bc Map) Cleanup() {
	n := time.Now()
	for k, v := range bc {
		if !v.Locked(n) {
			delete(bc, k)
		}
	}
}

func (bc Map) Backoff(key string) {
	c, ok := bc[key]
	if ok {
		bc[key] = Backoff(c)
	} else {
		bc[key] = Backoff()
	}
}

func (bc Map) Locked(key string, time time.Time) bool {
	c, ok := bc[key]
	if ok {
		return c.Locked(time)
	} else {
		return false
	}
}
