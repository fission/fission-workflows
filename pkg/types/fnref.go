package types

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	RuntimeDelimiter = "://"
	groupRuntime     = 1
	groupRuntimeID   = 2
)

var (
	fnRefReg        = regexp.MustCompile(fmt.Sprintf("^(?:(\\w+)%s)?(\\w+)$", RuntimeDelimiter))
	ErrInvalidFnRef = errors.New("invalid function reference")
	ErrNoRuntime    = errors.New("function reference does not contain a runtime")
	ErrNoRuntimeID  = errors.New("function reference does not contain a runtimeId")
)

func (m FnRef) Format() string {
	var runtime string
	if len(m.Runtime) > 0 {
		runtime = m.Runtime + RuntimeDelimiter
	}

	return runtime + m.ID
}

func (m FnRef) IsValid() bool {
	return IsFnRef(m.Format())
}

func (m FnRef) IsEmpty() bool {
	return m.ID == "" && m.Runtime == ""
}

func IsFnRef(s string) bool {
	_, err := ParseFnRef(s)
	return err == nil
}

func NewFnRef(runtime string, id string) FnRef {
	if len(id) == 0 {
		panic("function reference needs a runtime id")
	}
	return FnRef{
		Runtime: runtime,
		ID:      id,
	}
}

func ParseFnRef(s string) (FnRef, error) {
	matches := fnRefReg.FindStringSubmatch(s)
	if matches == nil {
		return FnRef{}, ErrInvalidFnRef
	}

	if len(matches[groupRuntimeID]) == 0 {
		return FnRef{
			Runtime: matches[groupRuntime],
		}, ErrNoRuntimeID
	}

	if len(matches[groupRuntime]) == 0 {
		return FnRef{
			ID: matches[groupRuntimeID],
		}, ErrNoRuntime
	}

	return FnRef{
		Runtime: matches[groupRuntime],
		ID:      matches[groupRuntimeID],
	}, nil
}
