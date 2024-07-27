package protoops

import (
	"iter"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// InterfaceOfListItem returns an interface value associated with the given list item value. It may panic or return invalid result if the field descriptor do not describes the list that the given item value belongs to. For scalar types and messages it returns its value.
func InterfaceOfListItem(listField protoreflect.FieldDescriptor, value protoreflect.Value) any {
	if listField.Kind() == protoreflect.MessageKind {
		return value.Message().Interface()
	}
	return value.Interface()
}

// SetMessageFieldListItem assigns the provided value to the list index inside a list field within the given message, by setting value under the specified index according to proto and protopatch assignment rules. It may panic if the provided parent message is invalid, the list field descriptor do not belongs to the message, list is not mutable or the given index is out of bounds of the list. It returns ErrMismatchingType error when value type do not matches any of types that allows assignment.
func SetMessageFieldListItem(parentMessage protoreflect.Message, listField protoreflect.FieldDescriptor, index int, to any) error {
	if to == nil {
		clearMessageFieldListItem(parentMessage, listField, index)
		return nil
	}
	if listField.Kind() == protoreflect.MessageKind {
		pr := ProtoreflectOfAny(to)
		if pr == nil {
			return ErrMismatchingType
		}
		if listField.Message() != pr.Descriptor() {
			return ErrMismatchingType
		}
		getList(parentMessage, listField).Set(index, protoreflect.ValueOfMessage(pr))
		return nil
	}
	if !IsTypeMatchesProtoScalarKind(listField.Kind(), reflect.TypeOf(to)) {
		return ErrMismatchingType
	}
	getList(parentMessage, listField).Set(index, protoreflect.ValueOf(to))
	return nil
}

// SetListItem assigns the provided value to the list index inside the given list, by setting value under the specified index according to proto and protopatch assignment rules. It may panic if the provided list is invalid or is read only, the given index is out of bounds of the list. It returns ErrMismatchingType error when value type do not matches any of types that allows assignment.
func SetListItem(li List, index int, to any) error {
	if to == nil {
		clearListItem(li, index)
		return nil
	}
	listField := li.ParentFieldDescriptor()
	if listField.Kind() == protoreflect.MessageKind {
		pr := ProtoreflectOfAny(to)
		if pr == nil {
			return ErrMismatchingType
		}
		if listField.Message() != pr.Descriptor() {
			return ErrMismatchingType
		}
		li.Set(index, protoreflect.ValueOfMessage(pr))
		return nil
	}
	if !IsTypeMatchesProtoScalarKind(listField.Kind(), reflect.TypeOf(to)) {
		return ErrMismatchingType
	}
	li.Set(index, protoreflect.ValueOf(to))
	return nil
}

func clearMessageFieldListItem(parentMessage protoreflect.Message, listField protoreflect.FieldDescriptor, index int) {
	clearListItem(getList(parentMessage, listField), index)
}

func clearListItem(li protoreflect.List, index int) {
	len := li.Len()
	for i := index + 1; i < len; i++ {
		li.Set(i-1, li.Get(i+1))
	}
	li.Truncate(len - 1)
}

// AppendMessageFieldListItem appends the provided value to end of the list value from a list field within the given message. It may panic if the provided parent message is invalid, the list field descriptor do not belongs to the message. It returns ErrMismatchingType error when value type do not matches any of types that would allow appending.
func AppendMessageFieldListItem(parentMessage protoreflect.Message, listField protoreflect.FieldDescriptor, new any) error {
	if new == nil {
		li := mutableList(parentMessage, listField)
		li.Append(li.NewElement())
		return nil
	}
	if listField.Kind() == protoreflect.MessageKind {
		pr := ProtoreflectOfAny(new)
		if pr == nil {
			return ErrMismatchingType
		}
		if listField.Message() != pr.Descriptor() {
			return ErrMismatchingType
		}
		mutableList(parentMessage, listField).Append(protoreflect.ValueOfMessage(pr))
		return nil
	}
	if !IsTypeMatchesProtoScalarKind(listField.Kind(), reflect.TypeOf(new)) {
		return ErrMismatchingType
	}
	mutableList(parentMessage, listField).Append(protoreflect.ValueOf(new))
	return nil
}

// AppendListItem appends the provided value to end of the provided list. It may panic if the provided list is invalid or is read only. It returns ErrMismatchingType error when value type do not matches any of types that would allow appending.
func AppendListItem(li List, new any) error {
	if new == nil {
		li.Append(li.NewElement())
		return nil
	}
	listField := li.ParentFieldDescriptor()
	if listField.Kind() == protoreflect.MessageKind {
		pr := ProtoreflectOfAny(new)
		if pr == nil {
			return ErrMismatchingType
		}
		if listField.Message() != pr.Descriptor() {
			return ErrMismatchingType
		}
		li.Append(protoreflect.ValueOfMessage(pr))
		return nil
	}
	if !IsTypeMatchesProtoScalarKind(listField.Kind(), reflect.TypeOf(new)) {
		return ErrMismatchingType
	}
	li.Append(protoreflect.ValueOf(new))
	return nil
}

// InsertMessageFieldListItem insert the provided value at specified index into the list value inside a list field within the given message, by expanding list, moving necessary items and setting value under the specified index according to proto and protopatch assignment rules. It may panic if the provided parent message is invalid, the list field descriptor do not belongs to the message or the given index is out of bounds of the list. It returns ErrMismatchingType error when value type do not matches any of types that would allow appending. Note thet when the index is exactly equal to the length of the list, insert behaves like append and do not panics due to out of bound index.
func InsertMessageFieldListItem(parentMessage protoreflect.Message, listField protoreflect.FieldDescriptor, index int, new any) error {
	if new == nil {
		li := mutableList(parentMessage, listField)
		insertListItemValue(li, index, li.NewElement())
		return nil
	}
	if listField.Kind() == protoreflect.MessageKind {
		pr := ProtoreflectOfAny(new)
		if pr == nil {
			return ErrMismatchingType
		}
		if listField.Message() != pr.Descriptor() {
			return ErrMismatchingType
		}
		insertListItemValue(mutableList(parentMessage, listField), index, protoreflect.ValueOfMessage(pr))
		return nil
	}
	if !IsTypeMatchesProtoScalarKind(listField.Kind(), reflect.TypeOf(new)) {
		return ErrMismatchingType
	}
	insertListItemValue(mutableList(parentMessage, listField), index, protoreflect.ValueOf(new))
	return nil
}

// InsertListItem insert the provided value at specified index into the provided list, by expanding list, moving necessary items and setting value under the specified index according to proto and protopatch assignment rules. It may panic if the provided list is invalid or is read only or the given index is out of bounds of the list. It returns ErrMismatchingType error when value type do not matches any of types that would allow appending. Note thet when the index is exactly equal to the length of the list, insert behaves like append and do not panics due to out of bound index.
func InsertListItem(li List, index int, new any) error {
	if new == nil {
		insertListItemValue(li, index, li.NewElement())
		return nil
	}
	listField := li.ParentFieldDescriptor()
	if listField.Kind() == protoreflect.MessageKind {
		pr := ProtoreflectOfAny(new)
		if pr == nil {
			return ErrMismatchingType
		}
		if listField.Message() != pr.Descriptor() {
			return ErrMismatchingType
		}
		insertListItemValue(li, index, protoreflect.ValueOfMessage(pr))
		return nil
	}
	if !IsTypeMatchesProtoScalarKind(listField.Kind(), reflect.TypeOf(new)) {
		return ErrMismatchingType
	}
	insertListItemValue(li, index, protoreflect.ValueOf(new))
	return nil
}

func insertListItemValue(li protoreflect.List, index int, new protoreflect.Value) {
	li.Append(new)
	len := li.Len()
	for i := index; i < len; i++ {
		v := li.Get(i)
		li.Set(i, new)
		new = v
	}
}

func getList(parentMessage protoreflect.Message, listField protoreflect.FieldDescriptor) protoreflect.List {
	return parentMessage.Get(listField).List()
}

func mutableList(parentMessage protoreflect.Message, listField protoreflect.FieldDescriptor) protoreflect.List {
	return parentMessage.Mutable(listField).List()
}

// List represents a protocol buffer list value.
type List interface {
	protoreflect.List
	Iter() iter.Seq2[int, protoreflect.Value]
	ParentFieldDescriptor() protoreflect.FieldDescriptor
	AsGoSlice() any
}

type listWrapper struct {
	field protoreflect.FieldDescriptor
	li    protoreflect.List
}

func NewList(parentField protoreflect.FieldDescriptor, li protoreflect.List) List {
	return listWrapper{field: parentField, li: li}
}

func (l listWrapper) Len() int                          { return l.li.Len() }
func (l listWrapper) Get(i int) protoreflect.Value      { return l.li.Get(i) }
func (l listWrapper) Set(i int, v protoreflect.Value)   { l.li.Set(i, v) }
func (l listWrapper) Append(v protoreflect.Value)       { l.li.Append(v) }
func (l listWrapper) AppendMutable() protoreflect.Value { return l.li.AppendMutable() }
func (l listWrapper) Truncate(n int)                    { l.li.Truncate(n) }
func (l listWrapper) NewElement() protoreflect.Value    { return l.li.NewElement() }
func (l listWrapper) IsValid() bool                     { return l.li.IsValid() }

func (l listWrapper) Iter() iter.Seq2[int, protoreflect.Value]            { return l.listRange }
func (l listWrapper) ParentFieldDescriptor() protoreflect.FieldDescriptor { return l.field }
func (l listWrapper) AsGoSlice() any                                      { return convertProtoreflectListToGoSlice(l.field, l.li) }

func (l listWrapper) listRange(yield func(int, protoreflect.Value) bool) {
	if !l.li.IsValid() {
		return
	}
	for i, len := 0, l.li.Len(); i < len; i++ {
		if !yield(i, l.li.Get(i)) {
			break
		}
	}
}

func convertProtoreflectListToGoSlice(field protoreflect.FieldDescriptor, li protoreflect.List) any {
	switch field.Kind() {
	case protoreflect.BoolKind:
		return convertProtoreflectListToGoSliceWithTypedValue[bool](li)
	case protoreflect.EnumKind:
		return convertProtoreflectListToGoSliceWithTypedValue[protoreflect.EnumNumber](li)
	case protoreflect.Int32Kind:
		return convertProtoreflectListToGoSliceWithTypedValue[int32](li)
	case protoreflect.Sint32Kind:
		return convertProtoreflectListToGoSliceWithTypedValue[int32](li)
	case protoreflect.Uint32Kind:
		return convertProtoreflectListToGoSliceWithTypedValue[uint32](li)
	case protoreflect.Int64Kind:
		return convertProtoreflectListToGoSliceWithTypedValue[int64](li)
	case protoreflect.Sint64Kind:
		return convertProtoreflectListToGoSliceWithTypedValue[int64](li)
	case protoreflect.Uint64Kind:
		return convertProtoreflectListToGoSliceWithTypedValue[uint64](li)
	case protoreflect.Sfixed32Kind:
		return convertProtoreflectListToGoSliceWithTypedValue[int32](li)
	case protoreflect.Fixed32Kind:
		return convertProtoreflectListToGoSliceWithTypedValue[uint32](li)
	case protoreflect.FloatKind:
		return convertProtoreflectListToGoSliceWithTypedValue[float32](li)
	case protoreflect.Sfixed64Kind:
		return convertProtoreflectListToGoSliceWithTypedValue[int64](li)
	case protoreflect.Fixed64Kind:
		return convertProtoreflectListToGoSliceWithTypedValue[uint64](li)
	case protoreflect.DoubleKind:
		return convertProtoreflectListToGoSliceWithTypedValue[float64](li)
	case protoreflect.StringKind:
		return convertProtoreflectListToGoSliceWithTypedValue[string](li)
	case protoreflect.BytesKind:
		return convertProtoreflectListToGoSliceWithTypedValue[[]byte](li)
	case protoreflect.MessageKind:
		return convertProtoreflectListOfMessagesToGoSlice(li)
	}
	panic("cannot convert protoreflect.List to Go slice: unsupported list type")
}

func convertProtoreflectListToGoSliceWithTypedValue[T any](li protoreflect.List) []T {
	s := make([]T, li.Len())
	for i := range s {
		s[i] = li.Get(i).Interface().(T)
	}
	return s
}

func convertProtoreflectListOfMessagesToGoSlice(li protoreflect.List) []proto.Message {
	s := make([]proto.Message, li.Len())
	for i := range s {
		s[i] = li.Get(i).Message().Interface()
	}
	return s
}
