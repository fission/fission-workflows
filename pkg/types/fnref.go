package types

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	RuntimeDelimiter = "://"
	groupRuntime     = 1
	groupRuntimeId   = 2
)

var (
	fnRefReg        = regexp.MustCompile(fmt.Sprintf("^(?:(\\w+)%s)?(\\w+)$", RuntimeDelimiter))
	ErrInvalidFnRef = errors.New("invalid function reference")
	ErrNoRuntime    = errors.New("function reference does not contain a runtime")
	ErrNoRuntimeId  = errors.New("function reference does not contain a runtimeId")
)

func (m FnRef) Format() string {
	var runtime string
	if len(m.Runtime) > 0 {
		runtime = m.Runtime + RuntimeDelimiter
	}

	return runtime + m.RuntimeId
}

func (m FnRef) IsValid() bool {
	return IsFnRef(m.Format())
}

func IsFnRef(s string) bool {
	_, err := ParseFnRef(s)
	return err == nil
}

func NewFnRef(runtime string, runtimeId string) FnRef {
	if len(runtimeId) == 0 {
		panic("function reference needs a runtime id")
	}
	return FnRef{
		Runtime:   runtime,
		RuntimeId: runtimeId,
	}
}

func ParseFnRef(s string) (FnRef, error) {
	matches := fnRefReg.FindStringSubmatch(s)
	if matches == nil {
		return FnRef{}, ErrInvalidFnRef
	}

	if len(matches[groupRuntimeId]) == 0 {
		return FnRef{
			Runtime: matches[groupRuntime],
		}, ErrNoRuntimeId
	}

	if len(matches[groupRuntime]) == 0 {
		return FnRef{
			RuntimeId: matches[groupRuntimeId],
		}, ErrNoRuntime
	}

	return FnRef{
		Runtime:   matches[groupRuntime],
		RuntimeId: matches[groupRuntimeId],
	}, nil
}
