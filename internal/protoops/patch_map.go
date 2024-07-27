package protoops

import (
	"iter"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// InterfaceOfMapItem returns an interface value associated with the given map item value. It may panic or return invalid result if the field descriptor do not describes the map that the given item value belongs to. For scalar types and messages it returns its value.
func InterfaceOfMapItem(mapField protoreflect.FieldDescriptor, value protoreflect.Value) any {
	if mapField.MapValue().Kind() == protoreflect.MessageKind {
		return value.Message().Interface()
	}
	return value.Interface()
}

// SetMessageFieldMapItem assigns the provided value to the map key inside a map field within the given message, by setting value under the specified map key according to proto and protopatch assignment rules. It may panic if the provided parent message is invalid, the map field descriptor does not describes a map field or does not belongs to the message, map key is invalid or its type do not matches map key type. It returns ErrMismatchingType error when value type do not matches any of types that allows assignment.
func SetMessageFieldMapItem(parentMessage protoreflect.Message, mapField protoreflect.FieldDescriptor, key protoreflect.MapKey, to any) error {
	if to == nil {
		clearMessageFieldMapItem(parentMessage, mapField, key)
		return nil
	}
	mapValue := mapField.MapValue()
	if mapValue.Kind() == protoreflect.MessageKind {
		pr := ProtoreflectOfAny(to)
		if pr == nil {
			return ErrMismatchingType
		}
		if mapValue.Message() != pr.Descriptor() {
			return ErrMismatchingType
		}
		mutableMap(parentMessage, mapField).Set(key, protoreflect.ValueOfMessage(pr))
		return nil
	}
	if !IsTypeMatchesProtoScalarKind(mapValue.Kind(), reflect.TypeOf(to)) {
		return ErrMismatchingType
	}
	mutableMap(parentMessage, mapField).Set(key, protoreflect.ValueOf(to))
	return nil
}

// SetMapItem assigns the provided value to the map key inside the given map, by setting value under the specified map key according to proto and protopatch assignment rules. It may panic if the provided map is invalid or read only, map key is invalid or its type do not matches map key type. It returns ErrMismatchingType error when value type do not matches any of types that allows assignment.
func SetMapItem(ma Map, key protoreflect.MapKey, to any) error {
	if to == nil {
		clearMapItem(ma, key)
		return nil
	}
	mapValue := ma.ParentFieldDescriptor().MapValue()
	if mapValue.Kind() == protoreflect.MessageKind {
		pr := ProtoreflectOfAny(to)
		if pr == nil {
			return ErrMismatchingType
		}
		if mapValue.Message() != pr.Descriptor() {
			return ErrMismatchingType
		}
		ma.Set(key, protoreflect.ValueOfMessage(pr))
		return nil
	}
	if !IsTypeMatchesProtoScalarKind(mapValue.Kind(), reflect.TypeOf(to)) {
		return ErrMismatchingType
	}
	ma.Set(key, protoreflect.ValueOf(to))
	return nil
}

func clearMessageFieldMapItem(parentMessage protoreflect.Message, mapField protoreflect.FieldDescriptor, key protoreflect.MapKey) {
	clearMapItem(getMap(parentMessage, mapField), key)
}

func clearMapItem(ma protoreflect.Map, key protoreflect.MapKey) {
	if !ma.IsValid() || ma.Len() == 0 {
		return
	}
	ma.Clear(key)
}

func getMap(parentMessage protoreflect.Message, mapField protoreflect.FieldDescriptor) protoreflect.Map {
	return parentMessage.Get(mapField).Map()
}

func mutableMap(parentMessage protoreflect.Message, mapField protoreflect.FieldDescriptor) protoreflect.Map {
	return parentMessage.Mutable(mapField).Map()
}

// Map represents a protocol buffer map value.
type Map interface {
	protoreflect.Map
	Iter() iter.Seq2[protoreflect.MapKey, protoreflect.Value]
	ParentFieldDescriptor() protoreflect.FieldDescriptor
	AsGoMap() any
}

type mapWrapper struct {
	field protoreflect.FieldDescriptor
	ma    protoreflect.Map
}

func NewMap(parentField protoreflect.FieldDescriptor, ma protoreflect.Map) Map {
	return mapWrapper{field: parentField, ma: ma}
}

func (m mapWrapper) Len() int                                                   { return m.ma.Len() }
func (m mapWrapper) Range(f func(protoreflect.MapKey, protoreflect.Value) bool) { m.ma.Range(f) }
func (m mapWrapper) Has(k protoreflect.MapKey) bool                             { return m.ma.Has(k) }
func (m mapWrapper) Clear(k protoreflect.MapKey)                                { m.ma.Clear(k) }
func (m mapWrapper) Get(k protoreflect.MapKey) protoreflect.Value               { return m.ma.Get(k) }
func (m mapWrapper) Set(k protoreflect.MapKey, v protoreflect.Value)            { m.ma.Set(k, v) }
func (m mapWrapper) Mutable(k protoreflect.MapKey) protoreflect.Value           { return m.ma.Mutable(k) }
func (m mapWrapper) NewValue() protoreflect.Value                               { return m.ma.NewValue() }
func (m mapWrapper) IsValid() bool                                              { return m.ma.IsValid() }

func (m mapWrapper) Iter() iter.Seq2[protoreflect.MapKey, protoreflect.Value] { return m.ma.Range }
func (m mapWrapper) ParentFieldDescriptor() protoreflect.FieldDescriptor      { return m.field }
func (m mapWrapper) AsGoMap() any                                             { return convertProtoreflectMapToGoMap(m.field, m.ma) }

func convertProtoreflectMapToGoMap(field protoreflect.FieldDescriptor, ma protoreflect.Map) any {
	switch field.MapKey().Kind() {
	case protoreflect.BoolKind:
		return convertProtoreflectMapToGoMapWithTypedKey[bool](field, ma)
	case protoreflect.Int32Kind:
		return convertProtoreflectMapToGoMapWithTypedKey[int32](field, ma)
	case protoreflect.Sint32Kind:
		return convertProtoreflectMapToGoMapWithTypedKey[int32](field, ma)
	case protoreflect.Uint32Kind:
		return convertProtoreflectMapToGoMapWithTypedKey[uint32](field, ma)
	case protoreflect.Int64Kind:
		return convertProtoreflectMapToGoMapWithTypedKey[int64](field, ma)
	case protoreflect.Sint64Kind:
		return convertProtoreflectMapToGoMapWithTypedKey[int64](field, ma)
	case protoreflect.Uint64Kind:
		return convertProtoreflectMapToGoMapWithTypedKey[uint64](field, ma)
	case protoreflect.Sfixed32Kind:
		return convertProtoreflectMapToGoMapWithTypedKey[int32](field, ma)
	case protoreflect.Fixed32Kind:
		return convertProtoreflectMapToGoMapWithTypedKey[uint32](field, ma)
	case protoreflect.Sfixed64Kind:
		return convertProtoreflectMapToGoMapWithTypedKey[int64](field, ma)
	case protoreflect.Fixed64Kind:
		return convertProtoreflectMapToGoMapWithTypedKey[uint64](field, ma)
	case protoreflect.StringKind:
		return convertProtoreflectMapToGoMapWithTypedKey[string](field, ma)
	}
	panic("cannot convert protoreflect.Map to Go map: unsupported map key type")
}

func convertProtoreflectMapToGoMapWithTypedKey[K comparable](field protoreflect.FieldDescriptor, ma protoreflect.Map) any {
	switch field.MapValue().Kind() {
	case protoreflect.BoolKind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, bool](ma)
	case protoreflect.EnumKind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, protoreflect.EnumNumber](ma)
	case protoreflect.Int32Kind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, int32](ma)
	case protoreflect.Sint32Kind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, int32](ma)
	case protoreflect.Uint32Kind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, uint32](ma)
	case protoreflect.Int64Kind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, int64](ma)
	case protoreflect.Sint64Kind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, int64](ma)
	case protoreflect.Uint64Kind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, uint64](ma)
	case protoreflect.Sfixed32Kind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, int32](ma)
	case protoreflect.Fixed32Kind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, uint32](ma)
	case protoreflect.FloatKind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, float32](ma)
	case protoreflect.Sfixed64Kind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, int64](ma)
	case protoreflect.Fixed64Kind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, uint64](ma)
	case protoreflect.DoubleKind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, float64](ma)
	case protoreflect.StringKind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, string](ma)
	case protoreflect.BytesKind:
		return convertProtoreflectMapToGoMapWithTypedKeyAndValue[K, []byte](ma)
	case protoreflect.MessageKind:
		return convertProtoreflectMapOfMessagesToGoMap[K](ma)
	}
	panic("cannot convert protoreflect.Map to Go map: unsupported map value type")
}

func convertProtoreflectMapToGoMapWithTypedKeyAndValue[K comparable, V any](ma protoreflect.Map) map[K]V {
	m := make(map[K]V, ma.Len())
	ma.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		m[k.Interface().(K)] = v.Interface().(V)
		return true
	})
	return m
}

func convertProtoreflectMapOfMessagesToGoMap[K comparable](ma protoreflect.Map) map[K]proto.Message {
	m := make(map[K]proto.Message, ma.Len())
	ma.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		m[k.Interface().(K)] = v.Message().Interface()
		return true
	})
	return m
}
