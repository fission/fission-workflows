package typedvalues

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/gogo/protobuf/proto"
)

const (
	TypeMap  = "map"
	TypeList = "list"
)

// TODO add iterator utility functions

func IsCollection(v ValueType) bool {
	return v == TypeMap || v == TypeList
}

type MapParserFormatter struct{}

func (fp *MapParserFormatter) Accepts() []ValueType {
	return []ValueType{
		TypeMap,
	}
}

func (fp *MapParserFormatter) Parse(ctx Parser, i interface{}) (*types.TypedValue, error) {
	var tvmp map[string]*types.TypedValue
	switch t := i.(type) {
	case map[string]interface{}:
		mp, err := ParseToTypedValueMap(ctx, t)
		if err != nil {
			return nil, err
		}
		tvmp = mp
	case map[string]*types.TypedValue:
		tvmp = t
	default:
		return nil, TypedValueErr{
			src: i,
			err: ErrUnsupportedType,
		}
	}
	return ParseTypedValueMap(tvmp)
}

func (fp *MapParserFormatter) Format(ctx Formatter, v *types.TypedValue) (interface{}, error) {
	mp, err := FormatToTypedValueMap(v)
	if err != nil {
		return nil, err
	}

	return FormatTypedValueMap(ctx, mp)
}

func ParseToTypedValueMap(ctx Parser, mp map[string]interface{}) (map[string]*types.TypedValue, error) {
	tvmp := map[string]*types.TypedValue{}
	for k, v := range mp {
		tvv, err := ctx.Parse(ctx, v)
		if err != nil {
			return nil, err
		}
		tvmp[k] = tvv
	}
	return tvmp, nil
}

func ParseTypedValueMap(mp map[string]*types.TypedValue) (*types.TypedValue, error) {
	pbv := &types.TypedValueMap{Value: mp}
	v, err := proto.Marshal(pbv)
	if err != nil {
		return nil, err
	}
	return &types.TypedValue{
		Type:  TypeMap,
		Value: v,
	}, nil
}

func FormatToTypedValueMap(v *types.TypedValue) (map[string]*types.TypedValue, error) {
	if v.Type != TypeMap {
		return nil, TypedValueErr{
			src: v,
			err: ErrUnsupportedType,
		}
	}

	mp := &types.TypedValueMap{}
	err := proto.Unmarshal(v.Value, mp)
	if err != nil {
		return nil, err
	}

	return mp.Value, nil
}

func FormatTypedValueMap(ctx Formatter, mp map[string]*types.TypedValue) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for k, v := range mp {
		tvv, err := ctx.Format(ctx, v)
		if err != nil {
			return nil, err
		}
		result[k] = tvv
	}
	return result, nil
}

type ListParserFormatter struct{}

func (pf *ListParserFormatter) Accepts() []ValueType {
	return []ValueType{
		TypeList,
	}
}

func (pf *ListParserFormatter) Parse(ctx Parser, i interface{}) (*types.TypedValue, error) {
	li, ok := i.([]interface{})
	if !ok {
		return nil, TypedValueErr{
			src: i,
			err: ErrUnsupportedType,
		}
	}

	return ParseList(ctx, li)
}

func (pf *ListParserFormatter) Format(ctx Formatter, v *types.TypedValue) (interface{}, error) {
	return FormatList(ctx, v)
}

func ParseList(ctx Parser, i []interface{}) (*types.TypedValue, error) {
	tvli, err := ParseListInterface(ctx, i)
	if err != nil {
		return nil, err
	}
	return ParseListTypedValue(tvli), nil
}

func ParseListInterface(ctx Parser, li []interface{}) ([]*types.TypedValue, error) {
	var result []*types.TypedValue
	for _, v := range li {
		tv, err := ctx.Parse(ctx, v)
		if err != nil {
			return nil, err
		}

		result = append(result, tv)
	}
	return result, nil
}

func ParseListTypedValue(l []*types.TypedValue) *types.TypedValue {
	tvl := &types.TypedValueList{Value: l}
	bs, err := proto.Marshal(tvl)
	if err != nil {
		panic(err)
	}

	return &types.TypedValue{
		Type:  TypeList,
		Value: bs,
	}
}

func FormatList(ctx Formatter, v *types.TypedValue) ([]interface{}, error) {
	ltv, err := FormatToTypedValueList(v)
	if err != nil {
		return nil, err
	}
	return FormatTypedValueList(ctx, ltv)
}

func FormatToTypedValueList(v *types.TypedValue) ([]*types.TypedValue, error) {
	err := verifyTypedValue(v, TypeList)
	if err != nil {
		return nil, err
	}

	tvl := &types.TypedValueList{}
	err = proto.Unmarshal(v.Value, tvl)
	if err != nil {
		return nil, err
	}
	return tvl.Value, nil
}

func FormatTypedValueList(ctx Formatter, vs []*types.TypedValue) ([]interface{}, error) {
	result := []interface{}{}
	for _, v := range vs {
		i, err := ctx.Format(ctx, v)
		if err != nil {
			return nil, err
		}
		result = append(result, i)
	}
	return result, nil
}
