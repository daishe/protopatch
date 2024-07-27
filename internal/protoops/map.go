package protoops

import (
	"reflect"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// IsTypeMapOfProtoMessageInterfaces reports weather the provided type is a map of proto message interfaces, that means it is map[K]proto.Message or map[K]protoreflace.Message where K is any valid map key type.
func IsTypeMapOfProtoMessageInterfaces(t reflect.Type) bool {
	if t == nil {
		return false
	}
	if t.Kind() != reflect.Map {
		return false
	}
	return isTypeProtoMessageInterface(t.Elem())
}

// MessageDescriptorFromValueOfMapOfProtoMessages returns common message descriptor form the provided map of proto message. If the given value does not contain a slice of proto messages or proto messages interfaces, if the there is more than one descriptor or if descriptor cannot be determined, function returns nil.
func MessageDescriptorFromValueOfMapOfProtoMessages(v reflect.Value) protoreflect.MessageDescriptor {
	if !v.IsValid() {
		return nil
	}
	if v.Kind() != reflect.Map {
		return nil
	}
	t := v.Type()
	if d := MessageDescriptorFromType(t.Elem()); d != nil {
		return d
	}
	if !IsTypeMapOfProtoMessageInterfaces(t) {
		return nil
	}
	first, d := true, protoreflect.MessageDescriptor(nil)
	for _, el := range v.Seq2() {
		if first {
			first, d = false, MessageDescriptorFromValue(el)
			continue
		}
		if d != MessageDescriptorFromValue(el) {
			return nil
		}
	}
	return d
}

// ParseMapKey parses the provided name as map key. It returns invalid map key if the provided name cannot be parsed.
func ParseMapKey(keyType protoreflect.FieldDescriptor, name string) protoreflect.MapKey {
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
	return k
}

// ParsedKeyInMap parses the provided name as map key. It returns invalid map key if the provided name cannot be parsed or was not found in map.
func ParsedKeyInMap(m protoreflect.Map, keyType protoreflect.FieldDescriptor, name string) protoreflect.MapKey {
	k := ParseMapKey(keyType, name)
	if !k.IsValid() || !m.Has(k) {
		return protoreflect.MapKey{}
	}
	return k
}
