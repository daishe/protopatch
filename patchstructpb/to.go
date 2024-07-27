package patchstructpb

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/daishe/protopatch"
	"github.com/daishe/protopatch/internal/protoops"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

// ToValueConverter returns a converter that attempts to convert the given proto type value into structpb.Value (or a value that is part of structpb.Value, like structpb.NullValue, structpb.Struct, structpb.ListValue, float64, string or bool).
func ToValueConverter(opts ...Option) protopatch.Converter {
	setup := newSetup(opts...)
	return protopatch.ConverterFunc(func(to, from any) (any, error) {
		return toAny(to, from, setup)
	})
}

// ConvertToValue attempts to convert the given proto type value into structpb.Value (or a value that is part of structpb.Value, like structpb.NullValue, structpb.Struct, structpb.ListValue, float64, string or bool).
func ConvertToValue(to, from any, opts ...Option) (any, error) {
	return toAny(to, from, newSetup(opts...))
}

func toAny(to, from any, setup *setup) (any, error) {
	switch to.(type) {
	case structpb.NullValue:
		v, err := toValue(from, setup)
		if err != nil {
			return nil, err
		}
		if _, ok := v.GetKind().(*structpb.Value_NullValue); ok {
			return structpb.NullValue_NULL_VALUE, nil
		}
	case bool:
		v, err := toValue(from, setup)
		if err != nil {
			return nil, err
		}
		if x, ok := v.GetKind().(*structpb.Value_BoolValue); ok {
			return x.BoolValue, nil
		}
	case float64:
		v, err := toValue(from, setup)
		if err != nil {
			return nil, err
		}
		if x, ok := v.GetKind().(*structpb.Value_NumberValue); ok {
			return x.NumberValue, nil
		}
		if x, ok := v.GetKind().(*structpb.Value_StringValue); ok {
			if f, err := strconv.ParseFloat(x.StringValue, 64); err == nil {
				return f, nil
			}
		}
	case string:
		v, err := toValue(from, setup)
		if err != nil {
			return nil, err
		}
		if x, ok := v.GetKind().(*structpb.Value_StringValue); ok {
			return x.StringValue, nil
		}
	case *structpb.Struct:
		v, err := toValue(from, setup)
		if err != nil {
			return nil, err
		}
		if x, ok := v.GetKind().(*structpb.Value_StructValue); ok {
			return x.StructValue, nil
		}
	case *structpb.ListValue:
		v, err := toValue(from, setup)
		if err != nil {
			return nil, err
		}
		if x, ok := v.GetKind().(*structpb.Value_ListValue); ok {
			return x.ListValue, nil
		}
	case *structpb.Value:
		v, err := toValue(from, setup)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, protopatch.ErrNoConversionDefined
}

func toValue(from any, setup *setup) (*structpb.Value, error) {
	switch v := from.(type) {
	case nil:
		return structpb.NewNullValue(), nil
	case structpb.NullValue:
		return structpb.NewNullValue(), nil
	case bool:
		return structpb.NewBoolValue(v), nil
	case int32:
		return convertIntToValue(int64(v)), nil
	case int64:
		return convertIntToValue(v), nil
	case uint32:
		return convertUintToValue(uint64(v)), nil
	case uint64:
		return convertUintToValue(v), nil
	case float32:
		return structpb.NewNumberValue(float64(v)), nil
	case float64:
		return structpb.NewNumberValue(v), nil
	case string:
		return structpb.NewStringValue(v), nil
	case []byte:
		return structpb.NewStringValue(string(v)), nil
	case proto.Message:
		return convertMessageToValue(v, setup)
	case protoreflect.Message:
		return convertMessageToValue(v.Interface(), setup)
	case protopatch.Map:
		return convertMapToValue(v, setup)
	case protopatch.List:
		return convertListToValue(v, setup)

	}
	v := reflect.ValueOf(from)
	t := v.Type()
	if t.Kind() == reflect.Slice {
		return convertGoSliceToValue(v, setup)
	}
	if t.Kind() == reflect.Map {
		return convertGoMapToValue(v, setup)
	}
	return nil, protopatch.ErrNoConversionDefined
}

func convertIntToValue(i int64) *structpb.Value {
	f := float64(i)
	if i == int64(f) {
		return structpb.NewNumberValue(f)
	}
	return structpb.NewStringValue(strconv.FormatInt(i, 10))
}

func convertUintToValue(i uint64) *structpb.Value {
	f := float64(i)
	if i == uint64(f) {
		return structpb.NewNumberValue(f)
	}
	return structpb.NewStringValue(strconv.FormatUint(i, 10))
}

func ConvertMessageToValue(m proto.Message, opts ...Option) (*structpb.Value, error) {
	return convertMessageToValue(m, newSetup(opts...))
}

func convertMessageToValue(m proto.Message, setup *setup) (*structpb.Value, error) {
	switch v := m.(type) {
	case *structpb.Struct:
		return structpb.NewStructValue(v), nil
	case *structpb.ListValue:
		return structpb.NewListValue(v), nil
	case *structpb.Value:
		return v, nil
	}
	// TODO: Support conversion when only descriptor matches any structpb message to support dynamic messages.

	pr := protoops.ProtoreflectOfMessage(m)
	if pr == nil {
		return nil, nil
	}
	if !pr.IsValid() {
		return structpb.NewStructValue(&structpb.Struct{}), nil
	}
	res := &structpb.Struct{Fields: map[string]*structpb.Value{}}
	dst := res.ProtoReflect()
	dstFields := dst.Descriptor().Fields().ByName("fields")
	fields := pr.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		if !pr.Has(f) {
			continue
		}
		conv, err := toValue(protoops.InterfaceOfMessageField(f, pr.Get(f)), setup)
		if err != nil {
			return nil, err
		}
		if err := protoops.SetMessageFieldMapItem(dst, dstFields, protoreflect.ValueOfString(f.JSONName()).MapKey(), conv); err != nil {
			return nil, err
		}
	}
	return structpb.NewStructValue(res), nil
}

func ConvertMapToValue(ma protopatch.Map, opts ...Option) (*structpb.Value, error) {
	return convertMapToValue(ma, newSetup(opts...))
}

func convertMapToValue(ma protopatch.Map, setup *setup) (*structpb.Value, error) {
	if ma == nil {
		return nil, nil
	}
	if !ma.IsValid() {
		return structpb.NewStructValue(&structpb.Struct{}), nil
	}
	res := &structpb.Struct{Fields: map[string]*structpb.Value{}}
	dst := res.ProtoReflect()
	dstFields := dst.Descriptor().Fields().ByName("fields")
	mapField := ma.ParentFieldDescriptor()
	for k, v := range ma.Iter() {
		conv, err := toValue(protoops.InterfaceOfMapItem(mapField, v), setup)
		if err != nil {
			return nil, err
		}
		if err := protoops.SetMessageFieldMapItem(dst, dstFields, protoreflect.ValueOfString(k.String()).MapKey(), conv); err != nil {
			return nil, err
		}
	}
	return structpb.NewStructValue(res), nil
}

func convertGoMapToValue(ma reflect.Value, setup *setup) (*structpb.Value, error) {
	res := &structpb.Struct{Fields: map[string]*structpb.Value{}}
	dst := res.ProtoReflect()
	dstFields := dst.Descriptor().Fields().ByName("fields")
	for k, v := range ma.Seq2() {
		dstKey, err := convertKeyToString(k.Interface())
		if err != nil {
			return nil, err
		}
		conv, err := toValue(v.Interface(), setup)
		if err != nil {
			return nil, err
		}
		if err := protoops.SetMessageFieldMapItem(dst, dstFields, protoreflect.ValueOfString(dstKey).MapKey(), conv); err != nil {
			return nil, err
		}
	}
	return structpb.NewStructValue(res), nil
}

func convertKeyToString(k any) (string, error) {
	switch v := k.(type) {
	case string:
		return v, nil
	case bool, int, int32, int64, uint, uint32, uint64:
		return fmt.Sprint(k), nil
	}
	return "", protopatch.ErrNoConversionDefined
}

func ConvertListToValue(li protopatch.List, opts ...Option) (*structpb.Value, error) {
	return convertListToValue(li, newSetup(opts...))
}

func convertListToValue(li protopatch.List, setup *setup) (*structpb.Value, error) {
	if li == nil {
		return nil, nil
	}
	if !li.IsValid() {
		return structpb.NewListValue(&structpb.ListValue{}), nil
	}
	res := &structpb.ListValue{}
	dst := res.ProtoReflect()
	dstFields := dst.Descriptor().Fields().ByName("values")
	listField := li.ParentFieldDescriptor()
	for _, v := range li.Iter() {
		conv, err := toValue(protoops.InterfaceOfListItem(listField, v), setup)
		if err != nil {
			return nil, err
		}
		if err := protoops.AppendMessageFieldListItem(dst, dstFields, conv); err != nil {
			return nil, err
		}
	}
	return structpb.NewListValue(res), nil
}

func convertGoSliceToValue(li reflect.Value, setup *setup) (*structpb.Value, error) {
	res := &structpb.ListValue{}
	dst := res.ProtoReflect()
	dstFields := dst.Descriptor().Fields().ByName("values")
	for _, v := range li.Seq2() {
		conv, err := toValue(v.Interface(), setup)
		if err != nil {
			return nil, err
		}
		if err := protoops.AppendMessageFieldListItem(dst, dstFields, conv); err != nil {
			return nil, err
		}
	}
	return structpb.NewListValue(res), nil
}
