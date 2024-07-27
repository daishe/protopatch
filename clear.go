package protopatch

import "google.golang.org/protobuf/proto"

func Clear(base proto.Message, path string, opts ...Option) error {
	return clearWithSetup(base, path, newSetup(opts...))
}

func clearWithSetup(base proto.Message, path string, setup *setup) error {
	if path == "" { // special case - an empty path; clear of the base message
		return clearSelf(base, setup)
	}

	c := MessageContainer(base)
	p := Path(path)

	if last := p.Last(); !last.IsFirst() { // path has more than 1 element
		a, err := access(c, last.PrecedingPath(), setup)
		if err != nil {
			return err
		}
		err = a.Set(last.Value(), nil)
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
	err = a.Set(path, nil)
	if err != nil {
		return err
	}
	return nil
}

func clearSelf(base proto.Message, setup *setup) error {
	c := MessageContainer(base).(*messageContainer)
	return c.setSelf(nil)
}

func (c *messageContainer) clearSelf() error {
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	fields := c.msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		c.msg.Clear(fields.Get(i))
	}
	return nil
}

func (c *messageContainer) clear(key string) error {
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	field, err := fieldInMessage(c.msg.Descriptor().Fields(), key)
	if err != nil {
		return err
	}
	c.msg.Clear(field)
	return nil
}

func (c *listContainer) clear(key string) error {
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	idx, err := indexInList(c.li, key)
	if err != nil {
		return err
	}
	len := c.li.Len()
	for i := idx + 1; i < len; i++ {
		c.li.Set(i-1, c.li.Get(i+1))
	}
	c.li.Truncate(len - 1)
	return nil
}

func (c *mapContainer) clear(key string) error {
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	mk, err := keyInMap(c.ma, c.parentField.MapKey(), key)
	if err != nil {
		return err
	}
	c.ma.Clear(mk)
	return nil
}

// func Clear(base protoreflect.Message, path string) error {
// 	err := clearInMessage(base, path)
// 	if r := recover(); r != nil {
// 		if err, ok := r.(error); ok {
// 			return fmt.Errorf("proto panic recovered: %w", NewErrInPath(path, err))
// 		}
// 		return fmt.Errorf("proto panic recovered: %w", NewErrInPath(path, fmt.Errorf("%v", r)))
// 	}
// 	return err
// }

// func clearInMessage(base protoreflect.Message, path string) error {
// 	if path == "" {
// 		fields := base.Descriptor().Fields()
// 		for i := 0; i < fields.Len(); i++ {
// 			base.Clear(fields.Get(i))
// 		}
// 		return nil
// 	}

// 	name, path := Cut(path)
// 	field, err := fieldInMessage(base.Descriptor().Fields(), name)
// 	if err != nil {
// 		return err
// 	}
// 	if path == "" {
// 		if base.Has(field) {
// 			base.Clear(field)
// 		}
// 		return nil
// 	}

// 	if field.IsList() {
// 		return NewErrInPath(name, clearInList(base, field, path))
// 	}
// 	if field.IsMap() {
// 		return NewErrInPath(name, clearInMap(base, field, path))
// 	}
// 	if field.Kind() == protoreflect.MessageKind {
// 		return NewErrInPath(name, clearInMessage(base.Get(field).Message(), path))
// 	}

// 	notFound, _ := Cut(path)
// 	return NewErrInPath(name, ErrNotFound{Kind: "field", Value: notFound})
// }

// func clearInList(base protoreflect.Message, listField protoreflect.FieldDescriptor, path string) error {
// 	if path == "" {
// 		if base.Has(listField) {
// 			base.Clear(listField)
// 		}
// 		return nil
// 	}

// 	name, path := Cut(path)
// 	if !base.Has(listField) {
// 		return ErrNotFound{Kind: "key", Value: name}
// 	}
// 	li := base.Mutable(listField).List()
// 	idx, err := indexInList(li, name)
// 	if err != nil {
// 		return err
// 	}
// 	if path == "" {
// 		li.Set(idx, li.NewElement())
// 		return nil
// 	}
// 	if listField.Kind() == protoreflect.MessageKind {
// 		return NewErrInPath(name, clearInMessage(li.Get(idx).Message(), path))
// 	}
// 	notFound, _ := Cut(path)
// 	return NewErrInPath(name, ErrNotFound{Kind: "field", Value: notFound})
// }

// func clearInMap(base protoreflect.Message, mapField protoreflect.FieldDescriptor, path string) error {
// 	mapValue := mapField.MapValue()
// 	if path == "" {
// 		if base.Has(mapField) {
// 			base.Clear(mapField)
// 		}
// 		return nil
// 	}

// 	name, path := Cut(path)
// 	if !base.Has(mapField) {
// 		return ErrNotFound{Kind: "key", Value: name}
// 	}
// 	m := base.Mutable(mapField).Map()
// 	key, err := keyInMap(m, mapField.MapKey(), name)
// 	if err != nil {
// 		return err
// 	}
// 	if path == "" {
// 		m.Set(key, m.NewValue())
// 		return nil
// 	}
// 	if mapValue.Kind() == protoreflect.MessageKind {
// 		return NewErrInPath(name, clearInMessage(m.Get(key).Message(), path))
// 	}
// 	notFound, _ := Cut(path)
// 	return NewErrInPath(name, ErrNotFound{Kind: "field", Value: notFound})
// }
