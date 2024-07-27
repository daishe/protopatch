package patchstructpb

import (
	"strconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/daishe/protopatch"
	"github.com/daishe/protopatch/internal/protoops"
)

// FromValueConverter returns a converter that attempts to convert structpb.Value (or a value that is part of structpb.Value, like structpb.NullValue, structpb.Struct, structpb.ListValue, float64, string or bool) to appropriate proto type.
func FromValueConverter(opts ...Option) protopatch.Converter {
	setup := newSetup(opts...)
	return protopatch.ConverterFunc(func(to, from any) (any, error) {
		return fromAny(to, from, setup)
	})
}

// ConvertFromValue attempts to convert structpb.Value (or a value that is part of structpb.Value, like structpb.NullValue, structpb.Struct, structpb.ListValue, float64, string or bool) to appropriate proto type.
func ConvertFromValue(to, from any, opts ...Option) (any, error) {
	return fromAny(to, from, newSetup(opts...))
}

func fromAny(to, from any, setup *setup) (any, error) {
	switch v := from.(type) {
	case structpb.NullValue:
		return nil, nil
	case float64:
		return fromNumberValue(to, v, setup)
	case string:
		return fromStringValue(to, v, setup)
	case bool:
		return fromBoolValue(to, v, setup)
	case *structpb.Struct:
		return fromStructValue(to, v, setup)
	case *structpb.ListValue:
		return fromListValue(to, v, setup)
	case *structpb.Value:
		return fromValue(to, v, setup)
	}
	// if setup.convertFromInterface {
	// 	switch v := from.(type) {
	// 	case map[string]any:
	// 		return fromStructValueInterface(to, v, setup)
	// 	case []any:
	// 		return fromListValueInterface(to, v, setup)
	// 	case nil:
	// 		return nil, nil
	// 	}
	// }
	return nil, protopatch.ErrNoConversionDefined
}

func fromValue(to any, from *structpb.Value, setup *setup) (conv any, err error) {
	switch k := from.GetKind().(type) {
	case *structpb.Value_NullValue:
		return nil, nil
	case *structpb.Value_NumberValue:
		return fromNumberValue(to, k.NumberValue, setup)
	case *structpb.Value_StringValue:
		return fromStringValue(to, k.StringValue, setup)
	case *structpb.Value_BoolValue:
		return fromBoolValue(to, k.BoolValue, setup)
	case *structpb.Value_StructValue:
		return fromStructValue(to, k.StructValue, setup)
	case *structpb.Value_ListValue:
		return fromListValue(to, k.ListValue, setup)
	default:
		return nil, nil // treat unset values as nulls
	}
}

func fromNumberValue(to any, from float64, _ *setup) (any, error) {
	switch to.(type) {
	case int32:
		if v := int32(from); from == float64(v) {
			return v, nil
		}
	case uint32:
		if v := uint32(from); from == float64(v) {
			return v, nil
		}
	case int64:
		if v := int64(from); from == float64(v) {
			return v, nil
		}
	case uint64:
		if v := uint64(from); from == float64(v) {
			return v, nil
		}
	case float32:
		return float32(from), nil
	case float64:
		return from, nil
	}
	return nil, protopatch.ErrNoConversionDefined
}

func fromStringValue(to any, from string, _ *setup) (any, error) {
	switch to.(type) {
	case string:
		return from, nil
	case []byte:
		return []byte(from), nil
	case int32:
		if v, err := strconv.ParseInt(from, 0, 32); err == nil {
			return int32(v), nil
		}
	case uint32:
		if v, err := strconv.ParseUint(from, 0, 32); err == nil {
			return uint32(v), nil
		}
	case int64:
		if v, err := strconv.ParseInt(from, 0, 64); err == nil {
			return int64(v), nil
		}
	case uint64:
		if v, err := strconv.ParseUint(from, 0, 64); err == nil {
			return uint64(v), nil
		}
	case float32:
		if v, err := strconv.ParseFloat(from, 32); err == nil {
			return float32(v), nil
		}
	case float64:
		if v, err := strconv.ParseFloat(from, 64); err == nil {
			return float64(v), nil
		}
	}
	return nil, protopatch.ErrNoConversionDefined
}

func fromBoolValue(to any, from bool, _ *setup) (any, error) {
	if _, ok := to.(bool); ok {
		return from, nil
	}
	return nil, protopatch.ErrNoConversionDefined
}

func fromStructValue(to any, from *structpb.Struct, setup *setup) (any, error) {
	if m, ok := to.(proto.Message); ok {
		return messageFromStructValue(m, from, setup)
	}
	if ma, ok := to.(protopatch.Map); ok {
		return mapFromStructValue(ma, from, setup)
	}
	return nil, protopatch.ErrNoConversionDefined
}

func fromListValue(to any, from *structpb.ListValue, setup *setup) (any, error) {
	if li, ok := to.(protopatch.List); ok {
		return listValueToList(li, from, setup)
	}
	return nil, protopatch.ErrNoConversionDefined
}

func messageFromStructValue(to proto.Message, from *structpb.Struct, setup *setup) (any, error) {
	pr := protoops.ProtoreflectOfMessage(to)
	if pr == nil {
		return nil, protopatch.ErrNoConversionDefined
	}
	desc := pr.Descriptor()
	if desc == (*structpb.Struct)(nil).ProtoReflect().Descriptor() {
		return from, nil
	}
	if desc == (*structpb.Value)(nil).ProtoReflect().Descriptor() {
		return structpb.NewStructValue(from), nil
	}
	for k, v := range from.GetFields() {
		f := protoops.FieldDescriptorInMessageDescriptor(desc, k)
		if f == nil {
			if setup.clearUnknownSourceStructKeys {
				delete(from.GetFields(), k)
				continue
			}
			if setup.ignoreUnknownStructKeysForMessages {
				continue
			}
			return nil, protopatch.ErrNoConversionDefined
		}
		conv, err := fromValue(protoops.InterfaceOfMessageField(f, pr.NewField(f)), v, setup)
		if err != nil {
			if setup.clearInvalidSourceValues {
				delete(from.GetFields(), k)
				continue
			}
			if setup.ignoreInvalidValues {
				continue
			}
			return nil, err
		}
		if err := protoops.SetMessageField(pr, f, conv); err != nil {
			if setup.clearInvalidSourceValues {
				delete(from.GetFields(), k)
				continue
			}
			if setup.ignoreInvalidValues {
				continue
			}
			return nil, protopatch.ErrNoConversionDefined
		}
	}
	return pr.Interface(), nil
}

func mapFromStructValue(to protopatch.Map, from *structpb.Struct, setup *setup) (any, error) {
	desc := to.ParentFieldDescriptor()
	structFieldsDesc := (*structpb.Struct)(nil).ProtoReflect().Descriptor().Fields().ByName("fields")
	if protoops.AreProtoFieldsMatch(desc, structFieldsDesc) {
		return from.GetFields(), nil
	}
	for k, v := range from.GetFields() {
		mk := protoops.ParseMapKey(desc.MapKey(), k)
		if !mk.IsValid() {
			if setup.clearUnknownSourceStructKeys {
				delete(from.GetFields(), k)
				continue
			}
			if setup.ignoreUnknownStructKeysForMaps {
				continue
			}
			return nil, protopatch.ErrNoConversionDefined
		}
		conv, err := fromValue(protoops.InterfaceOfMapItem(desc, to.NewValue()), v, setup)
		if err != nil {
			if setup.clearInvalidSourceValues {
				delete(from.GetFields(), k)
				continue
			}
			if setup.ignoreInvalidValues {
				continue
			}
			return nil, err
		}
		if conv == nil {
			conv = to.NewValue() // mimic list append behavior
		}
		if err := protoops.SetMapItem(to, mk, conv); err != nil {
			if setup.clearInvalidSourceValues {
				delete(from.GetFields(), k)
				continue
			}
			if setup.ignoreInvalidValues {
				continue
			}
			return nil, protopatch.ErrNoConversionDefined
		}
	}
	return to, nil
}

func listValueToList(to protopatch.List, from *structpb.ListValue, setup *setup) (any, error) {
	desc := to.ParentFieldDescriptor()
	listValuesDesc := (*structpb.ListValue)(nil).ProtoReflect().Descriptor().Fields().ByName("values")
	if protoops.AreProtoFieldsMatch(desc, listValuesDesc) {
		return from.GetValues(), nil
	}
	for i := 0; i < len(from.GetValues()); i++ {
		v := from.GetValues()[i]
		conv, err := fromValue(protoops.InterfaceOfListItem(desc, to.NewElement()), v, setup)
		if err != nil {
			if setup.clearInvalidSourceValues {
				from.Values = removeSliceIndex(from.GetValues(), i)
				i--
				continue
			}
			if setup.ignoreInvalidValues {
				continue
			}
			return nil, err
		}
		if err := protoops.AppendListItem(to, conv); err != nil {
			if setup.clearInvalidSourceValues {
				from.Values = removeSliceIndex(from.GetValues(), i)
				i--
				continue
			}
			if setup.ignoreInvalidValues {
				continue
			}
			return nil, protopatch.ErrNoConversionDefined
		}
	}
	return to, nil
}

func removeSliceIndex[T any](sl []T, idx int) []T {
	var zero T
	sl[idx] = zero
	return append(sl[:idx], sl[idx+1:]...)
}

func insertSliceIndex[T any](sl []T, idx int, new T) []T {
	sl = append(sl, new)
	len := len(sl)
	for i := idx; i < len; i++ {
		v := sl[i]
		sl[i] = new
		new = v
	}
	return sl
}
