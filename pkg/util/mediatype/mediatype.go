// Package mediatype implements the IANA Media Type standard.
//
// Although a partial implementation exists in the standard library under pkg/mime, it misses functionality, such as
// representation of suffices, and an explicit container format for media types. This package wraps the library,
// providing those functions.
//
// For Further information on media types see https://www.iana.org/assignments/media-types/media-types.xhtml
package mediatype

import (
	"mime"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

const (
	HeaderContentType = "Content-Type"
)

// See: https://en.wikipedia.org/wiki/Media_type
type MediaType struct {
	Parameters map[string]string
	Type       string
	Subtype    string
	Suffix     string
}

func (m *MediaType) Copy() *MediaType {
	if m == nil {
		return &MediaType{}
	}
	copiedParams := map[string]string{}
	for k, v := range m.Parameters {
		copiedParams[k] = v
	}
	return &MediaType{
		Parameters: copiedParams,
		Type:       m.Type,
		Subtype:    m.Subtype,
		Suffix:     m.Suffix,
	}
}

func (m *MediaType) ensureParametersExist() {
	if m.Parameters == nil {
		m.Parameters = map[string]string{}
	}
}

func (m *MediaType) Identifier() string {
	return m.Type + "/" + m.Subtype
}

func (m *MediaType) GetParam(key string) (string, bool) {
	m.ensureParametersExist()
	val, ok := m.Parameters[key]
	return val, ok
}

func (m *MediaType) SetParam(key string, val string) bool {
	m.ensureParametersExist()
	_, ok := m.Parameters[key]
	m.Parameters[key] = val
	return ok
}

func (m *MediaType) TypeEquals(other *MediaType) bool {
	if m == nil || other == nil {
		return m == other
	}

	return m.Type == other.Type && m.Subtype == other.Subtype
}

func (m *MediaType) String() string {
	if m == nil {
		return ""
	}
	var builder strings.Builder
	builder.WriteString(m.Type + "/" + m.Subtype)
	if len(m.Suffix) > 0 {
		builder.WriteString("+" + m.Suffix)
	}
	return mime.FormatMediaType(builder.String(), m.Parameters)
}

func SetContentTypeHeader(m *MediaType, w http.ResponseWriter) {
	w.Header().Set(HeaderContentType, m.String())
}

func Parse(s string) (*MediaType, error) {
	identifier, params, err := mime.ParseMediaType(s)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var mediaType, subType, suffix string
	if pos := strings.Index(identifier, "+"); pos >= 0 {
		suffix = identifier[pos+1:]
		identifier = identifier[:pos]
	}
	if pos := strings.Index(identifier, "/"); pos >= 0 {
		mediaType = identifier[:pos]
		subType = identifier[pos+1:]
	}
	return &MediaType{
		Parameters: params,
		Subtype:    subType,
		Type:       mediaType,
		Suffix:     suffix,
	}, nil
}

func MustParse(s string) *MediaType {
	mt, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return mt
}
