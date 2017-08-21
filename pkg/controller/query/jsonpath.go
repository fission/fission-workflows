package query

import (
	"fmt"
	"reflect"
	"strings"
)

/**
Use cases:
- Get task output
- Get workflow input
- Get task input
- Get task dependency
*/

/**

Supported functionality:
$ = from workflow
@ = current task
. = child

*/
// TODO return []interface by default to allow for collections
func Resolve(root interface{}, query string, cwd ...string) (interface{}, error) {
	// Decide starting node
	src := reflect.Indirect(reflect.ValueOf(root))
	switch query[0] {
	case '$': // src is correct
	case '@':
		if len(cwd) == 0 {
			return nil, fmt.Errorf("No cwd provided to resolve '%s'", query)
		}
		newSrc, err := traverse(src, cwd[0])
		if err != nil {
			return nil, err
		}
		src = reflect.Indirect(reflect.ValueOf(newSrc))
	default:
		return root, nil
	}

	// Iterate through selector
	return traverse(src, query)
}

// query: foo.bar
func traverse(root reflect.Value, query string) (interface{}, error) {
	path := strings.Split(query, ".")

	var result interface{}
	for _, node := range path {
		if node == "$" || node == "@" {
			result = root
			continue
		}

		field := root.FieldByName(node)
		if !field.IsValid() {
			return nil, nil
		}
		root = reflect.Indirect(field)
		result = root.Interface()
	}

	return result, nil
}
