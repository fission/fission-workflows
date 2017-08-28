package typedvalues

import (
	"encoding/json"

	"fmt"

	"github.com/fission/fission-workflow/pkg/types"
)

type Object map[string]interface{}

// The additional subtype (e.g. json/string) is needed to be able to evaluate inline functions on specific json types (e.g. strings vs. arrays)
const (
	FORMAT_JSON = "json"
	TYPE_STRING = "string"
	TYPE_OBJECT = "object"
	TYPE_ARRAY  = "array"
	TYPE_BOOL   = "bool"
)

var JSON_TYPES = []string{
	TYPE_STRING,
	TYPE_OBJECT,
	TYPE_ARRAY,
	TYPE_BOOL,
}

func isJsonValue(val *types.TypedValue) bool {
	for _, vtype := range JSON_TYPES {
		if FormatType(FORMAT_JSON, vtype) == val.Type {
			return true
		}
	}
	return false
}

type JsonParserFormatter struct{}

func (jp *JsonParserFormatter) Parse(i interface{}, allowedTypes ...string) (*types.TypedValue, error) {
	var tp string
	switch i.(type) {
	case bool:
		tp = TYPE_BOOL
	case string:
		tp = TYPE_STRING
	case map[string]interface{}:
		tp = TYPE_OBJECT
	case []interface{}:
		tp = TYPE_ARRAY
	default:
		return nil, fmt.Errorf("Value '%v' cannot be parsed to json", i)
	}

	bs, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	return &types.TypedValue{
		Type:  FormatType(FORMAT_JSON, tp),
		Value: bs,
	}, nil
}

func (jp *JsonParserFormatter) Format(v *types.TypedValue) (interface{}, error) {
	if !isJsonValue(v) {
		return nil, fmt.Errorf("Value '%v' is not a JSON type", v)
	}

	var i interface{}
	err := json.Unmarshal(v.Value, &i)
	return i, err
}
