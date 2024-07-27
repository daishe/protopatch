package protopatch

import (
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Insert(base proto.Message, path string, new any, opts ...Option) error {
	return insertWithSetup(base, path, new, newSetup(opts...))
}

func insertWithSetup(base proto.Message, path string, new any, setup *setup) error {
	if path == "" { // special case - an empty path; insert to the base message
		return ErrInsertToNonList
	}

	c := MessageContainer(base)
	p := Path(path)

	if last := p.Last(); !last.IsFirst() { // path has more than 1 element
		a, err := access(c, last.PrecedingPath(), setup)
		if err != nil {
			return err
		}
		ref, err := newElementForInsertInContainer(a)
		if err != nil {
			return NewErrInPath(string(last.PrecedingPath()), err)
		}
		conv, err := convert(ref, new, setup)
		if err != nil {
			return NewErrInPath(string(last.PrecedingPathWithCurrentSegment()), err)
		}
		err = a.Insert(last.Value(), conv)
		if err != nil {
			return NewErrInPath(string(last.PrecedingPath()), err)
		}
		return nil
	}

	// path has only 1 element
	a, err := transformContainer(c, setup)
	if err != nil {
		return err
	}
	ref, err := newElementForInsertInContainer(a)
	if err != nil {
		return err
	}
	conv, err := convert(ref, new, setup)
	if err != nil {
		return NewErrInPath(path, err)
	}
	err = c.Insert(path, conv)
	if err != nil {
		return err
	}
	return nil
}

func newElementForInsertInContainer(c Container) (any, error) {
	var kind protoreflect.Kind
	var val protoreflect.Value
	switch x := c.Self().(type) {
	case List:
		kind, val = x.ParentFieldDescriptor().Kind(), x.NewElement()
	case Map:
		kind, val = x.ParentFieldDescriptor().MapValue().Kind(), x.NewValue()
	default:
		return nil, ErrInsertToNonList
	}
	if kind == protoreflect.MessageKind {
		return val.Message().Interface(), nil
	}
	return val.Interface(), nil
}

func (c *messageContainer) Insert(key string, new any) error {
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	return ErrInsertToNonList
}

func (c *listContainer) Insert(key string, new any) error {
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	idx, err := indexInListForInsert(c.li, key)
	if err != nil {
		return err
	}
	if new == nil {
		c.insertCheckedValue(idx, c.li.NewElement())
		return nil
	}
	if c.parentField.Kind() == protoreflect.MessageKind {
		pr := asProtoreflectMessage(new)
		if pr == nil {
			return NewErrInPath(key, newInsertFailure(ErrMismatchingType))
		}
		if c.parentField.Message() != pr.Descriptor() {
			return NewErrInPath(key, newInsertFailure(ErrMismatchingType))
		}
		c.insertCheckedValue(idx, protoreflect.ValueOfMessage(pr))
		return nil
	}
	if reflect.TypeOf(c.li.NewElement().Interface()) != reflect.TypeOf(new) {
		return NewErrInPath(key, newInsertFailure(ErrMismatchingType))
	}
	c.insertCheckedValue(idx, protoreflect.ValueOf(new))
	return nil
}

func (c *listContainer) insertCheckedValue(idx int, new protoreflect.Value) {
	if !c.li.IsValid() {
		c.li = c.parent.Mutable(c.parentField).List()
	}
	c.li.Append(new)
	len := c.li.Len()
	for i := idx; i < len; i++ {
		v := c.li.Get(i)
		c.li.Set(i, new)
		new = v
	}
}

func (c *mapContainer) Insert(key string, new any) error {
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	return ErrInsertToNonList
	// mk, err := keyInMapForInsert(c.ma, c.parentField.MapKey(), key)
	// if err != nil {
	// 	return err
	// }
	// if new == nil {
	// 	c.ma.Set(mk, c.ma.NewValue())
	// 	return nil
	// }
	// if c.parentField.MapValue().Kind() == protoreflect.MessageKind {
	// 	pr := asProtoreflectMessage(new)
	// 	if pr == nil {
	// 		return NewErrInPath(key, newInsertFailure(ErrMismatchingType))
	// 	}
	// 	if c.parentField.MapValue().Message() != pr.Descriptor() {
	// 		return NewErrInPath(key, newInsertFailure(ErrMismatchingType))
	// 	}
	// 	c.setCheckedValue(mk, protoreflect.ValueOfMessage(pr))
	// 	return nil
	// }
	// ref := c.ma.Get(mk).Interface()
	// if !c.ma.Has(mk) {
	// 	ref = c.ma.NewValue().Interface()
	// }
	// if reflect.TypeOf(ref) != reflect.TypeOf(new) {
	// 	return NewErrInPath(key, newInsertFailure(ErrMismatchingType))
	// }
	// c.setCheckedValue(mk, protoreflect.ValueOf(new))
	// return nil
}

// func Insert(base protoreflect.Message, path string, value *structpb.Value) error {
// 	err := insertInMessage(base, path, value)
// 	if r := recover(); r != nil {
// 		if err, ok := r.(error); ok {
// 			return fmt.Errorf("proto panic recovered: %w", NewErrInPath(path, err))
// 		}
// 		return fmt.Errorf("proto panic recovered: %w", NewErrInPath(path, fmt.Errorf("%v", r)))
// 	}
// 	return err
// }

// func insertInMessage(base protoreflect.Message, path string, value *structpb.Value) error {
// 	if path == "" {
// 		return ErrInsertToNonList
// 	}

// 	name, path := Cut(path)
// 	field, err := fieldInMessage(base.Descriptor().Fields(), name)
// 	if err != nil {
// 		return err
// 	}
// 	if field.IsList() {
// 		return NewErrInPath(name, insertInList(base, field, path, value))
// 	}
// 	if field.IsMap() {
// 		return NewErrInPath(name, insertInMap(base, field, path, value))
// 	}
// 	if field.Kind() == protoreflect.MessageKind {
// 		// TOOD: Oneof check
// 		return NewErrInPath(name, insertInMessage(base.Mutable(field).Message(), path, value))
// 	}
// 	return NewErrInPath(name, ErrInsertToNonList)
// }

// func insertInList(base protoreflect.Message, listField protoreflect.FieldDescriptor, path string, value *structpb.Value) error {
// 	if path == "" { // append - add at the end
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
// 	if path != "" {
// 		if listField.Kind() == protoreflect.MessageKind {
// 			return NewErrInPath(name, insertInMessage(li.Get(idx).Message(), path, value))
// 		}
// 		notFound, _ := Cut(path)
// 		return NewErrInPath(name, ErrNotFound{Kind: "field", Value: notFound})
// 	}

// 	// insert - add at specific place
// 	v, err := structPbValueToListItem(li, listField, value)
// 	if err != nil {
// 		return NewErrInPath(name, err)
// 	}
// 	li.Append(v)

// 	for i := li.Len(); i >= idx; i-- {
// 		li.Set(i+1, li.Get(i))
// 	}
// 	li.Set(idx, v)
// 	return nil

// }

// func insertInMap(base protoreflect.Message, mapField protoreflect.FieldDescriptor, path string, value *structpb.Value) error {
// 	if path == "" {
// 		return ErrInsertToNonList
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
// 		return NewErrInPath(name, insertInMessage(m.Get(key).Message(), path, value))
// 	}
// 	if path != "" {
// 		notFound, _ := Cut(path)
// 		return NewErrInPath(name, ErrNotFound{Kind: "field", Value: notFound})
// 	}
// 	return NewErrInPath(name, ErrInsertToNonList)
// }
