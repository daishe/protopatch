package protopatch

import (
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Append(base proto.Message, path string, new any, opts ...Option) error {
	return appendWithSetup(base, path, new, newSetup(opts...))
}

func appendWithSetup(base proto.Message, path string, new any, setup *setup) error {
	if path == "" { // special case - an empty path; append to the base message
		return ErrAppendToNonList
	}
	p := Path(path)
	a, err := access(MessageContainer(base), p, setup)
	if err != nil {
		return err
	}
	ref, err := newElementForAppendInContainer(a)
	if err != nil {
		return NewErrInPath(string(p), err)
	}
	conv, err := convert(ref, new, setup)
	if err != nil {
		return NewErrInPath(string(p.Join("*")), err)
	}
	err = a.Append(conv)
	if err != nil {
		return NewErrInPath(string(p), err)
	}
	return nil
}

func newElementForAppendInContainer(c Container) (any, error) {
	li, ok := c.Self().(List)
	if !ok {
		return nil, ErrAppendToNonList
	}
	if li.ParentFieldDescriptor().Kind() == protoreflect.MessageKind {
		return li.NewElement().Message().Interface(), nil
	}
	return li.NewElement().Interface(), nil
}

func (c *messageContainer) Append(new any) error {
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	return ErrAppendToNonList
}

func (c *listContainer) Append(new any) error {
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	if c.parentField.Kind() == protoreflect.MessageKind {
		pr := asProtoreflectMessage(new)
		if pr == nil {
			return newAppendFailure(ErrMismatchingType)
		}
		if c.parentField.Message() != pr.Descriptor() {
			return newAppendFailure(ErrMismatchingType)
		}
		c.appendCheckedValue(protoreflect.ValueOfMessage(pr))
		return nil
	}
	if reflect.TypeOf(c.li.NewElement().Interface()) != reflect.TypeOf(new) {
		return newAppendFailure(ErrMismatchingType)
	}
	c.appendCheckedValue(protoreflect.ValueOf(new))
	return nil
}

func (c *listContainer) appendCheckedValue(new protoreflect.Value) {
	if !c.li.IsValid() {
		c.li = c.parent.Mutable(c.parentField).List()
	}
	c.li.Append(new)
}

func (c *mapContainer) Append(new any) error {
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	return ErrAppendToNonList
}

// func Append(base protoreflect.Message, path string, value *structpb.Value) error {
// 	err := appendInMessage(base, path, value)
// 	if r := recover(); r != nil {
// 		if err, ok := r.(error); ok {
// 			return fmt.Errorf("proto panic recovered: %w", NewErrInPath(path, err))
// 		}
// 		return fmt.Errorf("proto panic recovered: %w", NewErrInPath(path, fmt.Errorf("%v", r)))
// 	}
// 	return err
// }

// func appendInMessage(base protoreflect.Message, path string, value *structpb.Value) error {
// 	if path == "" {
// 		return ErrAppendToNonList
// 	}

// 	name, path := Cut(path)
// 	field, err := fieldInMessage(base.Descriptor().Fields(), name)
// 	if err != nil {
// 		return err
// 	}
// 	if field.IsList() {
// 		return NewErrInPath(name, appendInList(base, field, path, value))
// 	}
// 	if field.IsMap() {
// 		return NewErrInPath(name, appendInMap(base, field, path, value))
// 	}
// 	if field.Kind() == protoreflect.MessageKind {
// 		// TOOD: Oneof check
// 		return NewErrInPath(name, appendInMessage(base.Mutable(field).Message(), path, value))
// 	}
// 	return NewErrInPath(name, ErrAppendToNonList)
// }

// func appendInList(base protoreflect.Message, listField protoreflect.FieldDescriptor, path string, value *structpb.Value) error {
// 	if path == "" {
// 		li := base.Mutable(listField).List()
// 		v, err := structPbValueToListItem(li, listField, value)
// 		if err != nil {
// 			return err
// 		}
// 		li.Append(v)
// 		return nil
// 	}

// 	name, path := Cut(path)
// 	if !base.Has(listField) {
// 		return ErrNotFound{Kind: "field", Value: name}
// 	}
// 	li := base.Mutable(listField).List()
// 	idx, err := indexInList(li, name)
// 	if err != nil {
// 		return err
// 	}
// 	if listField.Kind() == protoreflect.MessageKind {
// 		return NewErrInPath(name, appendInMessage(li.Get(idx).Message(), path, value))
// 	}
// 	if path != "" {
// 		notFound, _ := Cut(path)
// 		return NewErrInPath(name, ErrNotFound{Kind: "field", Value: notFound})
// 	}
// 	return NewErrInPath(name, ErrAppendToNonList)

// }

// func appendInMap(base protoreflect.Message, mapField protoreflect.FieldDescriptor, path string, value *structpb.Value) error {
// 	if path == "" {
// 		return ErrAppendToNonList
// 	}

// 	name, path := Cut(path)
// 	if !base.Has(mapField) {
// 		return ErrNotFound{Kind: "field", Value: name}
// 	}
// 	m := base.Mutable(mapField).Map()
// 	key, err := keyInMap(m, mapField.MapKey(), name)
// 	if err != nil {
// 		return err
// 	}
// 	mapValue := mapField.MapValue()
// 	if mapValue.Kind() == protoreflect.MessageKind {
// 		return NewErrInPath(name, appendInMessage(m.Get(key).Message(), path, value))
// 	}
// 	if path != "" {
// 		notFound, _ := Cut(path)
// 		return NewErrInPath(name, ErrNotFound{Kind: "field", Value: notFound})
// 	}
// 	return NewErrInPath(name, ErrAppendToNonList)
// }
