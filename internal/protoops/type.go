package protoops

import (
	"reflect"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// InterfaceValue returns an protoreflect.Value associated with the given list item interface value. It may panic or return invalid result if the provided value cannot be represented as protoreflect.Value.
func InterfaceValue(i any) protoreflect.Value {
	if pr := ProtoreflectOfAny(i); pr != nil {
		return protoreflect.ValueOfMessage(pr)
	}
	return protoreflect.ValueOf(i)
}

// AreValueTypesMatch reports if types of two provided values match or if one can be set from the other using standard proto and protopatch rules for assignment.
func AreValueTypesMatch(x, y reflect.Value) bool {
	if !x.IsValid() || !y.IsValid() {
		return false
	}
	xType, yType := x.Type(), y.Type()
	if xType == yType {
		return true
	}
	if xDesc, yDesc := MessageDescriptorFromValue(x), MessageDescriptorFromValue(y); xDesc != nil && yDesc != nil && xDesc == yDesc {
		return true
	}
	if x.Kind() == reflect.Slice && y.Kind() == reflect.Slice {
		if xType.Elem() == yType.Elem() {
			return true
		}
		xDesc, yDesc := MessageDescriptorFromValueOfSliceOfProtoMessages(x), MessageDescriptorFromValueOfSliceOfProtoMessages(y)
		xIntWithZeroLen, yIntWithZeroLen := IsTypeSliceOfProtoMessageInterfaces(xType) && x.Len() == 0, IsTypeSliceOfProtoMessageInterfaces(yType) && y.Len() == 0
		if (xIntWithZeroLen && yIntWithZeroLen) || (xDesc != nil && yIntWithZeroLen) || (xIntWithZeroLen && yDesc != nil) || (xDesc != nil && yDesc != nil && xDesc == yDesc) {
			return true
		}
		return false
	}
	if x.Kind() == reflect.Map && y.Kind() == reflect.Map {
		if xType.Key() != yType.Key() {
			return false
		}
		if xType.Elem() == yType.Elem() {
			return true
		}
		xDesc, yDesc := MessageDescriptorFromValueOfMapOfProtoMessages(x), MessageDescriptorFromValueOfMapOfProtoMessages(y)
		xIntWithZeroLen, yIntWithZeroLen := IsTypeMapOfProtoMessageInterfaces(xType) && x.Len() == 0, IsTypeMapOfProtoMessageInterfaces(yType) && y.Len() == 0
		if (xIntWithZeroLen && yIntWithZeroLen) || (xDesc != nil && yIntWithZeroLen) || (xIntWithZeroLen && yDesc != nil) || (xDesc != nil && yDesc != nil && xDesc == yDesc) {
			return true
		}
		return false
	}
	return false
}

// IsValueTypeMatchesProtoField reports if type of the provided value matches the given protoreflect.FieldDescriptor. If the type matches it can be assigned to field using standard proto and protopatch rules for assignment.
func IsValueTypeMatchesProtoField(field protoreflect.FieldDescriptor, v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}
	t := v.Type()
	if field.IsList() {
		if t.Kind() != reflect.Slice {
			return false
		}
		if field.Kind() == protoreflect.MessageKind {
			if d := MessageDescriptorFromValueOfSliceOfProtoMessages(v); d != nil && field.Message() == d {
				return true
			}
			if IsTypeSliceOfProtoMessageInterfaces(t) && v.Len() == 0 {
				return true
			}
			return false
		}
		return IsTypeMatchesProtoScalarKind(field.Kind(), t.Elem())
	}
	if field.IsMap() {
		if t.Kind() != reflect.Map {
			return false
		}
		if !IsTypeMatchesProtoScalarKind(field.MapKey().Kind(), t.Key()) {
			return false
		}
		if field.MapValue().Kind() == protoreflect.MessageKind {
			if d := MessageDescriptorFromValueOfMapOfProtoMessages(v); d != nil && field.MapValue().Message() == d {
				return true
			}
			if IsTypeMapOfProtoMessageInterfaces(t) && v.Len() == 0 {
				return true
			}
			return false
		}
		return IsTypeMatchesProtoScalarKind(field.MapValue().Kind(), t.Elem())
	}
	if field.Kind() == protoreflect.MessageKind {
		return field.Message() == MessageDescriptorFromValue(v)
	}
	return IsTypeMatchesProtoScalarKind(field.Kind(), t.Elem())
}

// AreProtoFieldsMatch reports if tow provided protoreflect.FieldDescriptors match. If the field descriptors match, any of their value can be assigned to any of the fields using standard proto and protopatch rules for assignment.
func AreProtoFieldsMatch(x, y protoreflect.FieldDescriptor) bool {
	switch {
	case x.Kind() != y.Kind():
		return false

	// No enum descriptors check since protopatch allows to convert between enums freely.

	case x.Kind() == protoreflect.MessageKind && x.Message() != y.Message():
		return false

	case x.IsList() != y.IsList():
		return false

	case x.IsMap() != y.IsMap():
		return false
	case x.IsMap() && x.MapKey() != y.MapKey():
		return false
	case x.IsMap() && x.MapValue() != y.MapValue():
		return false
	}

	return true
}

// IsTypeMatchesProtoScalarKind reports is the provided protoreflect.Kind matches the given type. Check is only valid for scalar protoreflect types, that is: protoreflect.BoolKind, protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind, protoreflect.EnumKind, protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind, protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind, protoreflect.FloatKind, protoreflect.DoubleKind, protoreflect.StringKind, protoreflect.BytesKind.
func IsTypeMatchesProtoScalarKind(kind protoreflect.Kind, t reflect.Type) bool {
	if t == nil {
		return false
	}
	switch t.Kind() {
	case reflect.Bool:
		return kind == protoreflect.BoolKind
	case reflect.Int32:
		return kind == protoreflect.Int32Kind || kind == protoreflect.Sint32Kind || kind == protoreflect.Sfixed32Kind || kind == protoreflect.EnumKind
	case reflect.Int64:
		return kind == protoreflect.Int64Kind || kind == protoreflect.Sint64Kind || kind == protoreflect.Sfixed64Kind
	case reflect.Uint32:
		return kind == protoreflect.Uint32Kind || kind == protoreflect.Fixed32Kind
	case reflect.Uint64:
		return kind == protoreflect.Uint64Kind || kind == protoreflect.Fixed64Kind
	case reflect.Float32:
		return kind == protoreflect.FloatKind
	case reflect.Float64:
		return kind == protoreflect.DoubleKind
	case reflect.String:
		return kind == protoreflect.StringKind
	case reflect.Slice:
		return t.Elem().Kind() == reflect.Uint8 && kind == protoreflect.BytesKind
	}
	return false
}
