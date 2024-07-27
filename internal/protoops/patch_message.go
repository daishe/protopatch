package protoops

import (
	"reflect"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// InterfaceOfMessageField returns an interface value associated with the given field value. It may panic or return invalid result if the field descriptor do not describes the provided value. For scalar types and messages it returns its value. For lists and maps it returns List and Map interfaces accordingly.
func InterfaceOfMessageField(field protoreflect.FieldDescriptor, value protoreflect.Value) any {
	if field.IsList() {
		return NewList(field, value.List())
	}
	if field.IsMap() {
		return NewMap(field, value.Map())
	}
	if field.Kind() == protoreflect.MessageKind {
		return value.Message().Interface()
	}
	return value.Interface()
}

// SetMessageField assigns the provided value to message field, by setting field value according to proto and protopatch assignment rules. It may panic if the provided message is invalid or the field descriptor do not belongs to the message. It returns ErrMismatchingType error when value type do not matches any of types that allows assignment.
func SetMessageField(m protoreflect.Message, field protoreflect.FieldDescriptor, to any) error {
	if to == nil {
		clearMessageField(m, field)
		return nil
	}
	if field.IsList() {
		return setMessageFieldOfListType(m, field, to)
	}
	if field.IsMap() {
		return setMessageFieldOfMapType(m, field, to)
	}
	if field.Kind() == protoreflect.MessageKind {
		return setMessageFieldOfMessageType(m, field, to)
	}
	if !IsTypeMatchesProtoScalarKind(field.Kind(), reflect.TypeOf(to)) {
		return ErrMismatchingType
	}
	m.Set(field, protoreflect.ValueOf(to))
	return nil
}

func clearMessageField(m protoreflect.Message, field protoreflect.FieldDescriptor) {
	m.Clear(field)
}

func setMessageFieldOfMessageType(m protoreflect.Message, field protoreflect.FieldDescriptor, to any) error {
	pr := ProtoreflectOfAny(to)
	if pr == nil {
		return ErrMismatchingType
	}
	if field.Message() != pr.Descriptor() {
		return ErrMismatchingType
	}
	m.Set(field, protoreflect.ValueOfMessage(pr))
	return nil
}

func setMessageFieldOfListType(m protoreflect.Message, field protoreflect.FieldDescriptor, to any) error {
	if li, ok := to.(List); ok {
		to = li.AsGoSlice()
	}
	v := reflect.ValueOf(to)
	if !IsValueTypeMatchesProtoField(field, v) {
		return ErrMismatchingType
	}
	liVal := m.NewField(field)
	li := liVal.List()
	for _, i := range v.Seq2() {
		li.Append(InterfaceValue(i.Interface()))
	}
	m.Set(field, liVal)
	return nil
}

func setMessageFieldOfMapType(m protoreflect.Message, field protoreflect.FieldDescriptor, to any) error {
	if ma, ok := to.(Map); ok {
		to = ma.AsGoMap()
	}
	v := reflect.ValueOf(to)
	if !IsValueTypeMatchesProtoField(field, v) {
		return ErrMismatchingType
	}
	maVal := m.NewField(field)
	ma := maVal.Map()
	for k, el := range v.Seq2() {
		ma.Set(protoreflect.ValueOf(k.Interface()).MapKey(), InterfaceValue(el.Interface()))
	}
	m.Set(field, maVal)
	return nil
}
