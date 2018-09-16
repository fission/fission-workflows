package types

import (
	"errors"
	"net/url"
	"strings"
)

const (
	RuntimeDelimiter = "://"
)

var (
	ErrInvalidFnRef   = errors.New("invalid function reference")
	ErrFnRefNoRuntime = errors.New("fnref has empty runtime")
	ErrFnRefNoID      = errors.New("fnref has empty ID")
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
	return ValidateFnRef(m, false) == nil
}

func (m FnRef) IsEmpty() bool {
	return m.ID == "" && m.Runtime == "" && m.Namespace == ""
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

func ValidateFnRef(fnref FnRef, allowEmptyNamespace bool) error {
	if len(fnref.ID) == 0 {
		return ErrFnRefNoID
	}
	if !allowEmptyNamespace && len(fnref.Runtime) == 0 {
		return ErrFnRefNoRuntime
	}

	return nil
}

func ParseFnRef(s string) (FnRef, error) {
	u, err := url.Parse(s)
	if err != nil {
		return FnRef{}, ErrInvalidFnRef
	}
	scheme := u.Scheme
	ns := u.Host
	id := strings.Trim(u.Path, "/")
	if len(id) == 0 {
		id = strings.Trim(u.Host, "/")
		ns = ""
	}
	if len(id) == 0 {
		return FnRef{}, ErrInvalidFnRef
	}
	return FnRef{
		Runtime:   scheme,
		Namespace: ns,
		ID:        id,
	}, nil
}
