package protoops

import (
	"reflect"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// IsTypeSliceOfProtoMessageInterfaces reports weather the provided type is a slice of proto message interfaces, that means it is []proto.Message or []protoreflace.Message.
func IsTypeSliceOfProtoMessageInterfaces(t reflect.Type) bool {
	if t == nil {
		return false
	}
	if t.Kind() != reflect.Slice {
		return false
	}
	return isTypeProtoMessageInterface(t.Elem())
}

// MessageDescriptorFromValueOfSliceOfProtoMessages returns common message descriptor form the provided slice of proto message. If the given value does not contain a slice of proto messages or proto messages interfaces, if the there is more than one descriptor or if descriptor cannot be determined, function returns nil.
func MessageDescriptorFromValueOfSliceOfProtoMessages(v reflect.Value) protoreflect.MessageDescriptor {
	if !v.IsValid() {
		return nil
	}
	if v.Kind() != reflect.Slice {
		return nil
	}
	t := v.Type()
	if d := MessageDescriptorFromType(t.Elem()); d != nil {
		return d
	}
	if !IsTypeSliceOfProtoMessageInterfaces(t) {
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

// ParseListIndex parses the provided name as list index. It returns o and false if the provided name cannot be parsed. Note, the index can be negative, indicating reverse list counting.
func ParseListIndex(name string) (int, bool) {
	i, err := strconv.ParseInt(name, 10, 0)
	if err != nil {
		return 0, false
	}
	return int(i), true
}

// ParsedIndexInList parses the provided name as list index based on the given list length. It returns -1 if the provided name cannot be parsed or index is outside the bounds of list. Otherwise the function never returns negative indexes as negative indexes are normalized.
func ParsedIndexInList(len int, name string) int {
	i, ok := ParseListIndex(name)
	if !ok {
		return -1
	}
	if i < -len || i >= len {
		return -1
	}
	if i < 0 {
		i = len + i
	}
	return i
}
