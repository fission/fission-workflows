package typedvalues

import (
	"encoding/json"

	"fmt"

	"github.com/fission/fission-workflow/pkg/types"
)

type Object map[string]interface{}

// The additional subtype (e.g. json/string) is needed to be able to evaluate inline functions on specific json types (e.g. strings vs. arrays)
const (
	TYPE_REFERENCE = "ref"
	TYPE_STRING    = "json/string"
	TYPE_OBJECT    = "json/object"
)

var TYPES = []string{
	TYPE_STRING,
	TYPE_OBJECT,
}

func fromString(src string) (*types.TypedValue, error) {
	val, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	return &types.TypedValue{
		Type:  TYPE_STRING,
		Value: val,
	}, nil
}

func fromObject(src Object) (*types.TypedValue, error) {
	val, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	return &types.TypedValue{
		Type:  TYPE_OBJECT,
		Value: val,
	}, nil
}

func Supported(val *types.TypedValue) bool {
	for _, vtype := range TYPES {
		if vtype == val.Type {
			return true
		}
	}
	return false
}

func Parse(i interface{}) (*types.TypedValue, error) {
	// For now just duck-type conversions
	switch t := i.(type) {
	case string:
		return fromString(t)
	case Object:
		return fromObject(t)
	default:
		return nil, fmt.Errorf("Unknown type '%v' for i '%v'!", t, i)
	}
}

func From(t *types.TypedValue) (i interface{}) {
	json.Unmarshal(t.Value, &i)
	return i
}

// TODO extract to more abstract reference file
func Reference(ref string) *types.TypedValue {
	return &types.TypedValue{
		Type:  TYPE_REFERENCE,
		Value: []byte(ref),
	}
}

func Dereference(ref *types.TypedValue) string {
	return string(ref.Value)
}

func IsReference(value *types.TypedValue) bool {
	return value.Type == TYPE_REFERENCE
}
