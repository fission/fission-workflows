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
	ErrInvalidFnRef = errors.New("invalid function reference")
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
