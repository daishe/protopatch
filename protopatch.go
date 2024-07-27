package protopatch

import (
	"reflect"
	"strconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/daishe/protopatch/internal/protoops"
)

// List represents a protocol buffer list value.
type List = protoops.List

func NewList(parentField protoreflect.FieldDescriptor, li protoreflect.List) List {
	return protoops.NewList(parentField, li)
}

// Map represents a protocol buffer map value.
type Map = protoops.Map

func NewMap(parentField protoreflect.FieldDescriptor, ma protoreflect.Map) Map {
	return protoops.NewMap(parentField, ma)
}

// const PathSepearator = '.'

// func Join(segments ...string) (joined string) {
// 	for _, s := range segments {
// 		if joined != "" {
// 			joined += string(PathSepearator)
// 		}
// 		joined += s
// 	}
// 	return joined
// }

// func Cut(path string) (string, string) {
// 	first, path, _ := strings.Cut(path, string(PathSepearator))
// 	return first, path
// }

// func LastCut(path string) (string, string) {
// 	if i := strings.LastIndex(path, string(PathSepearator)); i >= 0 {
// 		return path[i+1:], path[:i]
// 	}
// 	return path, ""
// }

// func Split(path string) []string {
// 	if path == "" {
// 		return nil
// 	}
// 	return strings.Split(path, string(PathSepearator))
// }

func isTypeMatchesProtoScalarKind(kind protoreflect.Kind, typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Bool:
		return kind == protoreflect.BoolKind
	case reflect.Int32:
		return kind == protoreflect.Int32Kind || kind == protoreflect.Sint32Kind || kind == protoreflect.Sfixed32Kind
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
		return typ.Elem().Kind() == reflect.Uint8 && kind == protoreflect.BytesKind
	}
	return false
}

var (
	protoMessageType        = reflect.TypeOf((*proto.Message)(nil)).Elem()
	protoreflectMessageType = reflect.TypeOf((*protoreflect.Message)(nil)).Elem()
)

// func isTypeImplementationOfProtoMessage(msgDesc protoreflect.MessageDescriptor, typ reflect.Type) bool {
// 	pr := asProtoreflectMessage(reflect.Zero(typ).Interface())
// 	if pr == nil {
// 		return false
// 	}
// 	is := msgDesc == pr.Descriptor()
// 	if r := recover(); r != nil {
// 		return false
// 	}
// 	return is
// }

// func isValueImplementationOfProtoMessage(msgDesc protoreflect.MessageDescriptor, v reflect.Value) bool {
// 	if !v.IsValid() || v.Kind() != reflect.Pointer || v.IsNil() {
// 		return isTypeImplementationOfProtoMessage(msgDesc, v.Type())
// 	}
// 	pr := asProtoreflectMessage(v.Interface())
// 	if pr == nil {
// 		return false
// 	}
// 	is := msgDesc == pr.Descriptor()
// 	if r := recover(); r != nil {
// 		return false
// 	}
// 	return is
// }

func ignorePanic() bool {
	if r := recover(); r != nil {
		return true
	}
	return false
}

func asProtoreflectMessage(v any) protoreflect.Message {
	if pr, prOk := v.(protoreflect.Message); prOk {
		return pr
	}
	if m, mOk := v.(proto.Message); mOk {
		defer ignorePanic()
		pr := m.ProtoReflect()
		return pr
	}
	return nil
}

// func isSliceContainsImplementationsOfProtoMessage(msgDesc protoreflect.MessageDescriptor, v reflect.Value) bool {
// 	typ := v.Type()
// 	if typ.Kind() != reflect.Slice {
// 		return false
// 	}
// 	if typ.Elem() == protoMessageType {
// 		for el := range v.Seq() {
// 			m, ok := el.Interface().(proto.Message)
// 			if !ok {
// 				return false
// 			}
// 			pr := m.ProtoReflect()
// 			if pr == nil {
// 				return false
// 			}
// 			if pr.Descriptor() != msgDesc {
// 				return false
// 			}
// 		}
// 		return true
// 	}
// 	if typ.Elem() == protoreflectMessageType {
// 		for el := range v.Seq() {
// 			pr, ok := el.Interface().(protoreflect.Message)
// 			if !ok {
// 				return false
// 			}
// 			if pr.Descriptor() != msgDesc {
// 				return false
// 			}
// 		}
// 		return true
// 	}
// 	_ = recover()
// 	return false
// }

// func isMapContainsImplementationsOfProtoMessage(msgDesc protoreflect.MessageDescriptor, v reflect.Value) bool {
// 	typ := v.Type()
// 	if typ.Kind() != reflect.Map {
// 		return false
// 	}
// 	if typ.Elem() == protoMessageType {
// 		for _, el := range v.Seq2() {
// 			m, ok := el.Interface().(proto.Message)
// 			if !ok {
// 				return false
// 			}
// 			pr := m.ProtoReflect()
// 			if pr == nil {
// 				return false
// 			}
// 			if pr.Descriptor() != msgDesc {
// 				return false
// 			}
// 		}
// 		return true
// 	}
// 	if typ.Elem() == protoreflectMessageType {
// 		for _, el := range v.Seq2() {
// 			pr, ok := el.Interface().(protoreflect.Message)
// 			if !ok {
// 				return false
// 			}
// 			if pr.Descriptor() != msgDesc {
// 				return false
// 			}
// 		}
// 		return true
// 	}
// 	_ = recover()
// 	return false
// }

func isProtoMessageInterface(t reflect.Type) bool {
	return t == protoMessageType || t == protoreflectMessageType
}

func isSliceOfProtoMessageInterfaces(t reflect.Type) bool {
	if t.Kind() != reflect.Slice {
		return false
	}
	return isProtoMessageInterface(t.Elem())
}

func isMapOfProtoMessageInterfaces(t reflect.Type) bool {
	if t.Kind() != reflect.Map {
		return false
	}
	return isProtoMessageInterface(t.Elem())
}

func messageDescriptorFromType(typ reflect.Type) protoreflect.MessageDescriptor {
	pr := asProtoreflectMessage(reflect.Zero(typ).Interface())
	if pr == nil {
		return nil
	}
	defer ignorePanic()
	return pr.Descriptor()
}

func messageDescriptorFromValue(v reflect.Value) protoreflect.MessageDescriptor {
	if t := messageDescriptorFromType(v.Type()); t != nil {
		return t
	}
	if pr := asProtoreflectMessage(v.Interface()); pr != nil {
		defer ignorePanic()
		return pr.Descriptor()
	}
	return nil
}

func messageDescriptorOfSliceOfProtoMessages(v reflect.Value) protoreflect.MessageDescriptor {
	if v.Kind() != reflect.Slice {
		return nil
	}
	t := v.Type()
	if d := messageDescriptorFromType(t.Elem()); d != nil {
		return d
	}
	if !isSliceOfProtoMessageInterfaces(t) {
		return nil
	}
	first, d := true, protoreflect.MessageDescriptor(nil)
	for _, el := range v.Seq2() {
		if first {
			first, d = false, messageDescriptorFromValue(el)
			continue
		}
		if d != messageDescriptorFromValue(el) {
			return nil
		}
	}
	return d
}

func messageDescriptorOfMapOfProtoMessages(v reflect.Value) protoreflect.MessageDescriptor {
	if v.Kind() != reflect.Map {
		return nil
	}
	t := v.Type()
	if d := messageDescriptorFromType(t.Elem()); d != nil {
		return d
	}
	if !isMapOfProtoMessageInterfaces(t) {
		return nil
	}
	first, d := true, protoreflect.MessageDescriptor(nil)
	for _, el := range v.Seq2() {
		if first {
			first, d = false, messageDescriptorFromValue(el)
			continue
		}
		if d != messageDescriptorFromValue(el) {
			return nil
		}
	}
	return d
}

func areValueTypesMatch(x, y reflect.Value) bool {
	xType, yType := x.Type(), y.Type()
	if xType == yType {
		return true
	}
	if xDesc, yDesc := messageDescriptorFromValue(x), messageDescriptorFromValue(y); xDesc != nil && yDesc != nil && xDesc == yDesc {
		return true
	}
	if x.Kind() == reflect.Slice && y.Kind() == reflect.Slice {
		if areValueTypesMatch(x.Elem(), y.Elem()) {
			return true
		}
		if xDesc, yDesc := messageDescriptorOfSliceOfProtoMessages(x), messageDescriptorOfSliceOfProtoMessages(y); xDesc != nil && yDesc != nil && xDesc == yDesc {
			return true
		}
		if isSliceOfProtoMessageInterfaces(xType) && x.Len() == 0 && isSliceOfProtoMessageInterfaces(yType) && y.Len() == 0 {
			return true
		}
		return false
	}
	if x.Kind() == reflect.Map && y.Kind() == reflect.Map {
		if xType.Key() != yType.Key() {
			return false
		}
		if areValueTypesMatch(x.Elem(), y.Elem()) {
			return true
		}
		if xDesc, yDesc := messageDescriptorOfMapOfProtoMessages(x), messageDescriptorOfMapOfProtoMessages(y); xDesc != nil && yDesc != nil && xDesc == yDesc {
			return true
		}
		if isMapOfProtoMessageInterfaces(xType) && x.Len() == 0 && isMapOfProtoMessageInterfaces(yType) && y.Len() == 0 {
			return true
		}
		return false
	}
	return false
}

func isValueTypeMatchesProtoField(field protoreflect.FieldDescriptor, v reflect.Value) bool {
	t := v.Type()
	if field.IsList() {
		if t.Kind() != reflect.Slice {
			return false
		}
		if field.Kind() == protoreflect.MessageKind {
			if d := messageDescriptorOfSliceOfProtoMessages(v); d != nil && field.Message() == d {
				return true
			}
			if isSliceOfProtoMessageInterfaces(t) && v.Len() == 0 {
				return true
			}
			return false
		}
		return isTypeMatchesProtoScalarKind(field.Kind(), t.Elem())
	}
	if field.IsMap() {
		if t.Kind() != reflect.Map {
			return false
		}
		if !isTypeMatchesProtoScalarKind(field.MapKey().Kind(), t.Key()) {
			return false
		}
		if field.MapValue().Kind() == protoreflect.MessageKind {
			if d := messageDescriptorOfMapOfProtoMessages(v); d != nil && field.MapValue().Message() == d {
				return true
			}
			if isMapOfProtoMessageInterfaces(t) && v.Len() == 0 {
				return true
			}
			return false
		}
		return isTypeMatchesProtoScalarKind(field.MapValue().Kind(), t.Elem())
	}
	if field.Kind() == protoreflect.MessageKind {
		return field.Message() == messageDescriptorFromValue(v)
	}
	return isTypeMatchesProtoScalarKind(field.Kind(), t.Elem())
}

func areProtoFieldMatch(x, y protoreflect.FieldDescriptor) bool {
	switch {
	case x.Kind() != y.Kind():
		return false

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

func fieldInMessage(fields protoreflect.FieldDescriptors, name string) (protoreflect.FieldDescriptor, error) {
	if i, err := strconv.ParseInt(name, 10, 32); err == nil {
		return fields.ByNumber(protoreflect.FieldNumber(i)), nil
	}
	if field := fields.ByJSONName(name); field != nil {
		return field, nil
	}
	if field := fields.ByTextName(name); field != nil {
		return field, nil
	}
	if field := fields.ByName(protoreflect.Name(name)); field != nil {
		return field, nil
	}
	return nil, ErrNotFound{Kind: "field", Value: name}
}

func indexInList(list protoreflect.List, name string) (int, error) {
	len := list.Len()
	if len == 0 {
		return 0, ErrNotFound{Kind: "index", Value: name}
	}
	return indexUpToLen(len, name)
}

func indexInListForInsert(list protoreflect.List, name string) (int, error) {
	return indexUpToLen(list.Len()+1, name)
}

func indexUpToLen(len int, name string) (int, error) {
	i, err := parseListIndex(name)
	if err != nil {
		return 0, err
	}
	if i < -len || i >= len {
		return 0, ErrNotFound{Kind: "index", Value: name}
	}
	if i < 0 {
		i = len + i
	}
	return i, nil
}

func parseListIndex(name string) (int, error) {
	i, err := strconv.ParseInt(name, 10, 0)
	if err != nil {
		return 0, ErrNotFound{Kind: "index", Value: name}
	}
	return int(i), nil
}

func keyInMap(m protoreflect.Map, keyType protoreflect.FieldDescriptor, name string) (protoreflect.MapKey, error) {
	k, err := keyInMapForInsert(m, keyType, name)
	if err != nil {
		return protoreflect.MapKey{}, err
	}
	if !m.Has(k) {
		return protoreflect.MapKey{}, ErrNotFound{Kind: "key", Value: name}
	}
	return k, nil
}

func keyInMapForInsert(_ protoreflect.Map, keyType protoreflect.FieldDescriptor, name string) (protoreflect.MapKey, error) {
	return parseMapKey(keyType, name)
}

func parseMapKey(keyType protoreflect.FieldDescriptor, name string) (protoreflect.MapKey, error) {
	k := protoreflect.MapKey{}
	switch keyType.Kind() {
	case protoreflect.BoolKind:
		if v, err := strconv.ParseBool(name); err == nil {
			k = protoreflect.ValueOfBool(v).MapKey()
		}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		if v, err := strconv.ParseInt(name, 0, 32); err == nil {
			k = protoreflect.ValueOfInt32(int32(v)).MapKey()
		}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		if v, err := strconv.ParseInt(name, 0, 64); err == nil {
			k = protoreflect.ValueOfInt64(v).MapKey()
		}
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		if v, err := strconv.ParseUint(name, 0, 32); err == nil {
			k = protoreflect.ValueOfUint32(uint32(v)).MapKey()
		}
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		if v, err := strconv.ParseUint(name, 0, 64); err == nil {
			k = protoreflect.ValueOfUint64(v).MapKey()
		}
	case protoreflect.StringKind:
		k = protoreflect.ValueOfString(name).MapKey()
	}
	if !k.IsValid() {
		return protoreflect.MapKey{}, ErrNotFound{Kind: "key", Value: name}
	}
	return k, nil
}

// func structPbValueToMessage(to proto.Message, val *structpb.Value) error {
// 	b, err := protojson.Marshal(val)
// 	if err != nil {
// 		return err
// 	}
// 	if err := protojson.Unmarshal(b, to); err != nil {
// 		return fmt.Errorf("not a %s message value: %w", to.ProtoReflect().Descriptor().FullName(), err)
// 	}
// 	return nil
// }

// func structPbValueToScalar(field protoreflect.FieldDescriptor, val *structpb.Value) (protoreflect.Value, error) {
// 	switch field.Kind() {
// 	case protoreflect.BoolKind:
// 		if v, ok := val.Kind.(*structpb.Value_BoolValue); ok {
// 			return protoreflect.ValueOfBool(v.BoolValue), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not a bool value")

// 	case protoreflect.EnumKind:
// 		if v, ok := val.Kind.(*structpb.Value_StringValue); ok {
// 			ev := field.Enum().Values().ByName(protoreflect.Name(v.StringValue))
// 			if ev == nil {
// 				return protoreflect.Value{}, fmt.Errorf("no matching enum value for %q string", v.StringValue)
// 			}
// 			return protoreflect.ValueOfEnum(ev.Number()), nil
// 		}
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			ev := field.Enum().Values().ByNumber(protoreflect.EnumNumber(v.NumberValue))
// 			if ev == nil {
// 				return protoreflect.Value{}, fmt.Errorf("no matching enum value for %d number", protoreflect.EnumNumber(v.NumberValue))
// 			}
// 			return protoreflect.ValueOfEnum(ev.Number()), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not an enum value")

// 	case protoreflect.Int32Kind:
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			return protoreflect.ValueOfInt32(int32(v.NumberValue)), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not an int32 value")
// 	case protoreflect.Sint32Kind:
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			return protoreflect.ValueOfInt32(int32(v.NumberValue)), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not a sint32 value")
// 	case protoreflect.Sfixed32Kind:
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			return protoreflect.ValueOfInt32(int32(v.NumberValue)), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not a sfixed32 value")

// 	case protoreflect.Uint32Kind:
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			return protoreflect.ValueOfUint32(uint32(v.NumberValue)), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not an uint32 value")
// 	case protoreflect.Fixed32Kind:
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			return protoreflect.ValueOfUint32(uint32(v.NumberValue)), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not a fixed32 value")

// 	case protoreflect.Int64Kind:
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			return protoreflect.ValueOfInt64(int64(v.NumberValue)), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not an int64 value")
// 	case protoreflect.Sint64Kind:
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			return protoreflect.ValueOfInt64(int64(v.NumberValue)), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not a sint64 value")
// 	case protoreflect.Sfixed64Kind:
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			return protoreflect.ValueOfInt64(int64(v.NumberValue)), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not a sfixed64 value")

// 	case protoreflect.Uint64Kind:
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			return protoreflect.ValueOfUint64(uint64(v.NumberValue)), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not an uint64 value")
// 	case protoreflect.Fixed64Kind:
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			return protoreflect.ValueOfUint64(uint64(v.NumberValue)), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not a fixed64 value")

// 	case protoreflect.FloatKind:
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			return protoreflect.ValueOfFloat32(float32(v.NumberValue)), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not a float value")
// 	case protoreflect.DoubleKind:
// 		if v, ok := val.Kind.(*structpb.Value_NumberValue); ok {
// 			return protoreflect.ValueOfFloat64(v.NumberValue), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not a double value")

// 	case protoreflect.StringKind:
// 		if v, ok := val.Kind.(*structpb.Value_StringValue); ok {
// 			return protoreflect.ValueOfString(v.StringValue), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not a string value")

// 	case protoreflect.BytesKind:
// 		if v, ok := val.Kind.(*structpb.Value_StringValue); ok {
// 			return protoreflect.ValueOfBytes([]byte(v.StringValue)), nil
// 		}
// 		return protoreflect.Value{}, fmt.Errorf("not a bytes value")
// 	}

// 	return protoreflect.Value{}, fmt.Errorf("kind %s is not a scalar value", field.Kind())
// }

// func structPbValueToList(base protoreflect.Message, listField protoreflect.FieldDescriptor, val *structpb.Value) error {
// 	// TODO: oneof check
// 	wrapped := structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
// 		listField.JSONName(): val,
// 	}})
// 	b, err := protojson.Marshal(wrapped)
// 	if err != nil {
// 		return err
// 	}
// 	m := base.New()
// 	if err := protojson.Unmarshal(b, m.Interface()); err != nil {
// 		if listField.Kind() == protoreflect.MessageKind {
// 			return fmt.Errorf("not a %s message list value: %w", listField.Message().FullName(), err)
// 		}
// 		if listField.Kind() == protoreflect.EnumKind {
// 			return fmt.Errorf("not a %s enum list value: %w", listField.Enum().FullName(), err)
// 		}
// 		return fmt.Errorf("not a %s list value: %w", listField.Kind().String(), err)
// 	}
// 	base.Set(listField, m.Get(listField))
// 	return nil
// }

// func structPbValueToMap(base protoreflect.Message, mapField protoreflect.FieldDescriptor, val *structpb.Value) error {
// 	// TODO: oneof check
// 	wrapped := structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
// 		mapField.JSONName(): val,
// 	}})
// 	b, err := protojson.Marshal(wrapped)
// 	if err != nil {
// 		return err
// 	}
// 	m := base.New()
// 	if err := protojson.Unmarshal(b, m.Interface()); err != nil {
// 		if mapField.Kind() == protoreflect.MessageKind {
// 			return fmt.Errorf("not a %s message map value: %w", mapField.Message().FullName(), err)
// 		}
// 		if mapField.Kind() == protoreflect.EnumKind {
// 			return fmt.Errorf("not a %s enum map value: %w", mapField.Enum().FullName(), err)
// 		}
// 		return fmt.Errorf("not a %s map value: %w", mapField.Kind().String(), err)
// 	}
// 	base.Set(mapField, m.Get(mapField))
// 	return nil
// }

// func structPbValueToListItem(li protoreflect.List, listField protoreflect.FieldDescriptor, val *structpb.Value) (protoreflect.Value, error) {
// 	if listField.Kind() == protoreflect.MessageKind {
// 		v := li.NewElement()
// 		if err := structPbValueToMessage(v.Message().Interface(), val); err != nil {
// 			return protoreflect.Value{}, err
// 		}
// 		return v, nil
// 	}
// 	return structPbValueToScalar(listField, val)
// }
