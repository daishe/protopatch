package protoops

import (
	"reflect"
	"strconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ProtoreflectOfAny returns protoreflect.Message value of the provided proto.Message or protoreflect.Message. If the provided value does not contains a valid proto.Message or protoreflect.Message function returns nil.
func ProtoreflectOfAny(v any) protoreflect.Message {
	if pr, prOk := v.(protoreflect.Message); prOk {
		return pr
	}
	if m, mOk := v.(proto.Message); mOk {
		return ProtoreflectOfMessage(m)
	}
	return nil
}

// ProtoreflectOfMessage returns protoreflect.Message value of the provided proto.Message or nil if the message is invalid.
func ProtoreflectOfMessage(m proto.Message) protoreflect.Message {
	if m == nil {
		return nil
	}
	return m.ProtoReflect()
}

// NewProtoreflectOfMessage returns new mutable protoreflect.Message value having a type like the provided proto.Message or nil if the message is invalid.
func NewProtoreflectOfMessage(m proto.Message) protoreflect.Message {
	pr := ProtoreflectOfMessage(m)
	if pr == nil {
		return nil
	}
	if !pr.IsValid() {
		return pr.Type().New()
	}
	return pr.New()
}

var (
	protoMessageType        = reflect.TypeOf((*proto.Message)(nil)).Elem()
	protoreflectMessageType = reflect.TypeOf((*protoreflect.Message)(nil)).Elem()
)

// isTypeProtoMessageInterface reports weather the provided type is a proto message interface, that means it is proto.Message or protoreflace.Message.
func isTypeProtoMessageInterface(t reflect.Type) bool {
	return t == protoMessageType || t == protoreflectMessageType
}

// MessageDescriptorFromType retrieves protoreflect.MessageDescriptor if the provided type contains an implementation of proto.Message or protoreflect.Message. Otherwise it returns nil.
func MessageDescriptorFromType(t reflect.Type) protoreflect.MessageDescriptor {
	if t == nil {
		return nil
	}
	m, ok := reflect.Zero(t).Interface().(proto.Message)
	if !ok {
		return nil
	}
	pr := ProtoreflectOfMessage(m)
	if pr == nil {
		return nil
	}
	return pr.Descriptor()
}

// MessageDescriptorFromType retrieves protoreflect.MessageDescriptor if the provided value contains an implementation of proto.Message or protoreflect.Message. Otherwise it returns nil.
func MessageDescriptorFromValue(v reflect.Value) protoreflect.MessageDescriptor {
	if !v.IsValid() {
		return nil
	}
	if t := MessageDescriptorFromType(v.Type()); t != nil {
		return t
	}
	if pr := ProtoreflectOfAny(v.Interface()); pr != nil {
		return pr.Descriptor()
	}
	return nil
}

// FieldDescriptorInMessageDescriptor parses the provided name as field name. It returns nil field descriptor if the associated field cannot be found.
func FieldDescriptorInMessageDescriptor(desc protoreflect.MessageDescriptor, name string) protoreflect.FieldDescriptor {
	if desc == nil {
		return nil
	}
	fields := desc.Fields()
	if field := fields.ByJSONName(name); field != nil {
		return field
	}
	if field := fields.ByTextName(name); field != nil {
		return field
	}
	if field := fields.ByName(protoreflect.Name(name)); field != nil {
		return field
	}
	if i, err := strconv.ParseInt(name, 10, 32); err == nil {
		return fields.ByNumber(protoreflect.FieldNumber(i))
	}
	return nil
}
