package types

import (
	"errors"
	"fmt"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RuntimeDelimiter = "://"
)

const (
	// These are the index of capture groups for the FnRef regex
	groupFnRef = iota
	groupRuntime
	groupNamespace
	groupRuntimeID
)

var (
	fnRefReg        = regexp.MustCompile(fmt.Sprintf("^(?:(\\w+)%s)?(?:([a-zA-Z0-9][a-zA-Z0-9_-]{1,128})/)?(\\w+)$", RuntimeDelimiter))
	ErrInvalidFnRef = errors.New("invalid function reference")
	ErrNoRuntime    = errors.New("function reference does not contain a runtime")
	ErrNoRuntimeID  = errors.New("function reference does not contain a runtimeId")
)

func (m FnRef) Format() string {
	var runtime string
	if len(m.Runtime) > 0 {
		runtime = m.Runtime + RuntimeDelimiter
	}

	if len(m.Namespace) > 0 {
		return runtime + m.Namespace + `/` + m.ID

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

func NewFnRef(runtime, ns, id string) FnRef {
	if len(id) == 0 {
		panic("function reference needs a runtime id")
	}
	return FnRef{
		Runtime:   runtime,
		Namespace: ns,
		ID:        id,
	}
}

func ParseFnRef(s string) (FnRef, error) {
	matches := fnRefReg.FindStringSubmatch(s)
	if matches == nil {
		return FnRef{}, ErrInvalidFnRef
	}

	ns := matches[groupNamespace]
	if len(ns) == 0 {
		ns = metav1.NamespaceDefault
	}

	return FnRef{
		Runtime:   matches[groupRuntime],
		Namespace: ns,
		ID:        matches[groupRuntimeID],
	}, nil
}
