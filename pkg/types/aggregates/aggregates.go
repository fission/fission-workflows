package aggregates

import (
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/golang/protobuf/ptypes"
)

func unmarshalEventData(event *fes.Event) (interface{}, error) {
	d := &ptypes.DynamicAny{}
	err := ptypes.UnmarshalAny(event.Data, d)
	if err != nil {
		return nil, err
	}
	return d.Message, nil
}
