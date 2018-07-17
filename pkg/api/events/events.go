package events

import (
	"reflect"

	"github.com/golang/protobuf/proto"
)

func TypeOf(event proto.Message) string {
	if event == nil {
		return ""
	}
	return reflect.Indirect(reflect.ValueOf(event)).Type().Name()
}
