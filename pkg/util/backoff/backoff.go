package backoff

import (
	"math"
	"time"
)

// Utility for managing backoff strategies.
// Future: add a consistent log formatter
// Future: store err and all info per attempt (to allow this to function as the general retry library)

type Algorithm interface {
	Backoff(context ...Context) Context
}

var DefaultBackoffAlgorithm = &ExponentialBackoff{
	MaxBackoff: time.Duration(1) * time.Hour,
	Exponent:   2,
	Step:       time.Duration(100) * time.Millisecond,
}

func Backoff(context ...Context) Context {
	return DefaultBackoffAlgorithm.Backoff(context...)
}

type Context struct {
	LockedUntil time.Time
	Attempts    int
	Algorithm   Algorithm
	Lockout     time.Duration
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
	return bc.LockedUntil.After(time)
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
		l = context[0].LockedUntil
	}

	max := eb.MaxBackoff.Nanoseconds()
	min := eb.MinBackoff.Nanoseconds()
	step := eb.Step.Nanoseconds()
	exp := eb.Exponent
	if max == 0 {
		max = math.MaxInt64
	}

	d := (0.5 * (math.Pow(float64(exp), float64(c)) - 1.0)) * float64(step)

	lockoutDuration := math.Max(math.Min(d, float64(max)), float64(min))

	lockout := time.Duration(lockoutDuration) * time.Nanosecond
	return Context{
		Algorithm:   eb,
		Attempts:    c + 1,
		LockedUntil: l.Add(lockout),
		Lockout:     lockout,
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

func (bc Map) Reset(key string) {
	delete(bc, key)
}

func (bc Map) Backoff(key string) Context {
	c, ok := bc[key]
	var nb Context
	if ok {
		nb = Backoff(c)
	} else {
		nb = Backoff()
	}
	bc[key] = nb
	return nb
}

func (bc Map) Locked(key string, time time.Time) bool {
	c, ok := bc[key]
	if ok {
		return c.Locked(time)
	} else {
		return false
	}
}
