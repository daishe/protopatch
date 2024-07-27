package protopatch

import (
	"errors"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// func mapIfWellKnownValue(v Value) Value {
// 	return v // TODO: do mapping with well known types
// }

// func Get(container Container, path Path) (any, error) {
// 	err := error(nil)
// 	last := path.Last()
// 	if !last.IsFirst() {
// 		if container, err = Access(container, last.PrecedingPath()); err != nil {
// 			return nil, err
// 		}
// 	}
// 	value, err := container.Get(last.Value())
// 	if err != nil {
// 		if !last.IsFirst() {
// 			return nil, NewErrInPath(string(last.PrecedingPath()), err)
// 		}
// 		return nil, err
// 	}
// 	return value, err
// }

// func Mutable(container Container, path Path) (any, error) {
// 	err := error(nil)
// 	last := path.Last()
// 	if !last.IsFirst() {
// 		if container, err = AccessMutable(container, last.PrecedingPath()); err != nil {
// 			return nil, err
// 		}
// 	}
// 	value, err := container.Mutable(last.Value())
// 	if err != nil {
// 		if !last.IsFirst() {
// 			return nil, NewErrInPath(string(last.PrecedingPath()), err)
// 		}
// 		return nil, err
// 	}
// 	return value, err
// }

func Access(container Container, path Path, opts ...Option) (Container, error) {
	return access(container, path, newSetup(opts...))
}

func access(container Container, path Path, setup *setup) (Container, error) {
	container, err := transformContainer(container, setup)
	if err != nil {
		return nil, err
	}
	for ps := range path.Iter {
		next, err := accessOnce(container, ps, setup)
		if err != nil {
			return nil, err
		}
		container = next
	}
	return container, nil
}

func accessOnce(c Container, ps PathSegment, setup *setup) (Container, error) {
	next, err := c.Access(ps.Value())
	if err != nil {
		if errors.Is(err, ErrAccessToNonContainer) {
			return nil, NewErrInPath(string(ps.PrecedingPathWithCurrentSegment()), err)
		}
		if !ps.IsFirst() {
			return nil, NewErrInPath(string(ps.PrecedingPath()), err)
		}
		return nil, err
	}
	next, err = transformContainer(next, setup)
	if err != nil {
		return nil, NewErrInPath(string(ps.PrecedingPathWithCurrentSegment()), err)
	}
	return next, nil
}

func AccessMutable(container Container, path Path, opts ...Option) (Container, error) {
	return accessMutable(container, path, newSetup(opts...))
}

func accessMutable(container Container, path Path, setup *setup) (Container, error) {
	container, err := transformContainer(container, setup)
	if err != nil {
		return nil, err
	}
	for ps := range path.Iter {
		next, err := accessMutableOnce(container, ps, setup)
		if err != nil {
			return nil, err
		}
		container = next
	}
	return container, nil
}

func accessMutableOnce(c Container, ps PathSegment, setup *setup) (Container, error) {
	next, err := c.AccessMutable(ps.Value())
	if err != nil {
		if errors.Is(err, ErrAccessToNonContainer) {
			return nil, NewErrInPath(string(ps.PrecedingPathWithCurrentSegment()), err)
		}
		if !ps.IsFirst() {
			return nil, NewErrInPath(string(ps.PrecedingPath()), err)
		}
		return nil, err
	}
	next, err = transformContainer(next, setup)
	if err != nil {
		return nil, NewErrInPath(string(ps.PrecedingPathWithCurrentSegment()), err)
	}
	return next, nil
}

func (c *messageContainer) Self() any {
	return c.msg.Interface()
}

func (c *messageContainer) Get(key string) (any, error) {
	field, err := fieldInMessage(c.msg.Descriptor().Fields(), key)
	if err != nil {
		return nil, err
	}
	if field.IsList() {
		return NewList(field, c.msg.Get(field).List()), nil
	}
	if field.IsMap() {
		return NewMap(field, c.msg.Get(field).Map()), nil
	}
	if field.Kind() == protoreflect.MessageKind {
		return c.msg.Get(field).Message().Interface(), nil
	}
	return c.msg.Get(field).Interface(), nil
}

func (c *messageContainer) GetCopy(key string) (any, error) {
	field, err := fieldInMessage(c.msg.Descriptor().Fields(), key)
	if err != nil {
		return nil, err
	}
	if field.IsList() {
		li := NewList(field, c.msg.NewField(field).List())
		copyList(NewList(field, c.msg.Get(field).List()), li)
		return li, nil
	}
	if field.IsMap() {
		ma := NewMap(field, c.msg.NewField(field).Map())
		copyMap(NewMap(field, c.msg.Get(field).Map()), ma)
		return ma, nil
	}
	if field.Kind() == protoreflect.MessageKind {
		return proto.Clone(c.msg.Get(field).Message().Interface()), nil
	}
	return c.msg.Get(field).Interface(), nil
}

func copyList(src, dst List) {
	for _, v := range src.Iter() {
		dst.Append(v)
	}
}

func copyMap(src, dst Map) {
	for k, v := range src.Iter() {
		dst.Set(k, v)
	}
}

func (c *messageContainer) GetNew(key string) (any, error) {
	field, err := fieldInMessage(c.msg.Descriptor().Fields(), key)
	if err != nil {
		return nil, err
	}
	if field.IsList() {
		return NewList(field, c.msg.NewField(field).List()), nil
	}
	if field.IsMap() {
		return NewMap(field, c.msg.NewField(field).Map()), nil
	}
	if field.Kind() == protoreflect.MessageKind {
		return proto.Clone(c.msg.NewField(field).Message().Interface()), nil
	}
	return c.msg.Get(field).Interface(), nil
}

func (c *messageContainer) Mutable(key string) (any, error) {
	if c.ro {
		return nil, ErrMutationOfReadOnlyValue
	}
	field, err := fieldInMessage(c.msg.Descriptor().Fields(), key)
	if err != nil {
		return nil, err
	}
	if field.IsList() {
		return NewList(field, c.msg.Mutable(field).List()), nil
	}
	if field.IsMap() {
		return NewMap(field, c.msg.Mutable(field).Map()), nil
	}
	if field.Kind() == protoreflect.MessageKind {
		return c.msg.Mutable(field).Message().Interface(), nil
	}
	if field.HasPresence() && !c.msg.Has(field) {
		c.msg.Set(field, c.msg.NewField(field))
	}
	return c.msg.Get(field).Interface(), nil
}

func (c *messageContainer) Access(key string) (Container, error) {
	field, err := fieldInMessage(c.msg.Descriptor().Fields(), key)
	if err != nil {
		return nil, err
	}
	ro := c.ro || (field.HasPresence() && !c.msg.Has(field))
	if field.IsList() {
		return newListContainer(c.msg, field, c.msg.Get(field).List(), ro), nil
	}
	if field.IsMap() {
		return newMapContainer(c.msg, field, c.msg.Get(field).Map(), ro), nil
	}
	if field.Kind() == protoreflect.MessageKind {
		return newMessageContainer(c.msg.Get(field).Message(), ro), nil
	}
	return nil, ErrAccessToNonContainer
}

func (c *messageContainer) AccessMutable(key string) (Container, error) {
	if c.ro {
		return nil, ErrMutationOfReadOnlyValue
	}
	field, err := fieldInMessage(c.msg.Descriptor().Fields(), key)
	if err != nil {
		return nil, err
	}
	if field.IsList() {
		return newListContainer(c.msg, field, c.msg.Mutable(field).List(), false), nil
	}
	if field.IsMap() {
		return newMapContainer(c.msg, field, c.msg.Mutable(field).Map(), false), nil
	}
	if field.Kind() == protoreflect.MessageKind {
		return newMessageContainer(c.msg.Mutable(field).Message(), false), nil
	}
	if field.HasPresence() && !c.msg.Has(field) {
		c.msg.Set(field, c.msg.NewField(field))
	}
	return nil, ErrAccessToNonContainer
}

// func (v *messageFieldValue) AccessReadOnly(name string) (Value, error) {
// 	if v.field.IsList() {
// 		if v.ro || !v.msg.Has(v.field) {
// 			return nil, ErrFieldNotFound{Field: name}
// 		}
// 		li := v.msg.Get(v.field).List()
// 		idx, err := indexInList(li, name)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return mapIfWellKnownValue(newListElementValue(li, idx)), nil
// 	}
// 	if v.field.IsMap() {
// 		if v.ro || !v.msg.Has(v.field) {
// 			return nil, ErrFieldNotFound{Field: name}
// 		}
// 		ma := v.msg.Get(v.field).Map()
// 		key, err := keyInMap(ma, name)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return mapIfWellKnownValue(newMapElementValue(ma, key)), nil
// 	}
// 	if v.field.Kind() == protoreflect.MessageKind {
// 		return newMessageContainer(v.msg.Get(v.field).Message(), v.ro || !v.msg.Has(v.field)).AccessReadOnly(name)
// 	}
// 	return nil, ErrFieldNotFound{Field: name}
// }

// func (v *messageFieldValue) AccessMutable(name string) (Value, error) {
// 	if v.ro {
// 		return nil, ErrMutationOfReadOnlyValue
// 	}
// 	if v.field.IsList() {
// 		li := v.msg.Mutable(v.field).List()
// 		idx, err := indexInList(li, name)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return mapIfWellKnownValue(newListElementValue(li, idx)), nil
// 	}
// 	if v.field.IsMap() {
// 		ma := v.msg.Mutable(v.field).Map()
// 		key, err := keyInMap(ma, name)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return mapIfWellKnownValue(newMapElementValue(ma, key)), nil
// 	}
// 	if v.field.Kind() == protoreflect.MessageKind {
// 		return newMessageContainer(v.msg.Mutable(v.field).Message(), false).AccessMutable(name)
// 	}
// 	return nil, ErrFieldNotFound{Field: name}
// }

func (c *listContainer) Self() any {
	return NewList(c.parentField, c.li)
}

func (c *listContainer) Get(key string) (any, error) {
	idx, err := indexInList(c.li, key)
	if err != nil {
		return nil, err
	}
	if c.parentField.Kind() == protoreflect.MessageKind {
		return c.li.Get(idx).Message().Interface(), nil
	}
	return c.li.Get(idx).Interface(), nil
}

func (c *listContainer) GetCopy(key string) (any, error) {
	idx, err := indexInList(c.li, key)
	if err != nil {
		return nil, err
	}
	if c.parentField.Kind() == protoreflect.MessageKind {
		return proto.Clone(c.li.Get(idx).Message().Interface()), nil
	}
	return c.li.Get(idx).Interface(), nil
}

func (c *listContainer) GetNew(key string) (any, error) {
	_, err := parseListIndex(key)
	if err != nil {
		return nil, err
	}
	if c.parentField.Kind() == protoreflect.MessageKind {
		return c.li.NewElement().Message().Interface(), nil
	}
	return c.li.NewElement().Interface(), nil
}

func (c *listContainer) Mutable(key string) (any, error) {
	if c.ro {
		return nil, ErrMutationOfReadOnlyValue
	}
	idx, err := indexInList(c.li, key)
	if err != nil {
		return nil, err
	}
	if c.parentField.Kind() == protoreflect.MessageKind {
		return c.li.Get(idx).Message().Interface(), nil
	}
	return c.li.Get(idx).Interface(), nil
}

func (c *listContainer) Access(key string) (Container, error) {
	idx, err := indexInList(c.li, key)
	if err != nil {
		return nil, err
	}
	if c.parentField.Kind() != protoreflect.MessageKind {
		return nil, ErrAccessToNonContainer
	}
	return newMessageContainer(c.li.Get(idx).Message(), c.ro), nil
}

func (c *listContainer) AccessMutable(key string) (Container, error) {
	if c.ro {
		return nil, ErrMutationOfReadOnlyValue
	}
	idx, err := indexInList(c.li, key)
	if err != nil {
		return nil, err
	}
	if c.parentField.Kind() != protoreflect.MessageKind {
		return nil, ErrAccessToNonContainer
	}
	return newMessageContainer(c.li.Get(idx).Message(), c.ro), nil
}

// func (v *listElementValue) AccessReadOnly(name string) (Value, error) {
// 	if _, ok := v.li.NewElement().Interface().(proto.Message); !ok {
// 		return nil, ErrKeyNotFound{Key: name}
// 	}
// 	return newMessageContainer(v.li.Get(v.idx).Message(), false).AccessReadOnly(name)
// }

// func (v *listElementValue) AccessMutable(name string) (Value, error) {
// 	if _, ok := v.li.NewElement().Interface().(proto.Message); !ok {
// 		return nil, ErrKeyNotFound{Key: name}
// 	}
// 	return newMessageContainer(v.li.Get(v.idx).Message(), false).AccessMutable(name)
// }

func (c *mapContainer) Self() any {
	return NewMap(c.parentField, c.ma)
}

func (c *mapContainer) Get(key string) (any, error) {
	mk, err := keyInMap(c.ma, c.parentField.MapKey(), key)
	if err != nil {
		return nil, err
	}
	if c.parentField.MapValue().Kind() == protoreflect.MessageKind {
		return c.ma.Get(mk).Message().Interface(), nil
	}
	return c.ma.Get(mk).Interface(), nil
}

func (c *mapContainer) GetCopy(key string) (any, error) {
	mk, err := keyInMap(c.ma, c.parentField.MapKey(), key)
	if err != nil {
		return nil, err
	}
	if c.parentField.MapValue().Kind() == protoreflect.MessageKind {
		return proto.Clone(c.ma.Get(mk).Message().Interface()), nil
	}
	return c.ma.Get(mk).Interface(), nil
}

func (c *mapContainer) GetNew(key string) (any, error) {
	_, err := parseMapKey(c.parentField.MapKey(), key)
	if err != nil {
		return nil, err
	}
	if c.parentField.MapValue().Kind() == protoreflect.MessageKind {
		return c.ma.NewValue().Message().Interface(), nil
	}
	return c.ma.NewValue().Interface(), nil
}

func (c *mapContainer) Mutable(key string) (any, error) {
	if c.ro {
		return nil, ErrMutationOfReadOnlyValue
	}
	mk, err := keyInMap(c.ma, c.parentField.MapKey(), key)
	if err != nil {
		return nil, err
	}
	if c.parentField.MapValue().Kind() == protoreflect.MessageKind {
		return c.ma.Get(mk).Message().Interface(), nil
	}
	return c.ma.Get(mk).Interface(), nil
}

func (c *mapContainer) Access(key string) (Container, error) {
	mk, err := keyInMap(c.ma, c.parentField.MapKey(), key)
	if err != nil {
		return nil, err
	}
	if c.parentField.MapValue().Kind() != protoreflect.MessageKind {
		return nil, ErrAccessToNonContainer
	}
	return newMessageContainer(c.ma.Get(mk).Message(), c.ro), nil
}

func (c *mapContainer) AccessMutable(key string) (Container, error) {
	if c.ro {
		return nil, ErrMutationOfReadOnlyValue
	}
	mk, err := keyInMap(c.ma, c.parentField.MapKey(), key)
	if err != nil {
		return nil, err
	}
	if c.parentField.MapValue().Kind() != protoreflect.MessageKind {
		return nil, ErrAccessToNonContainer
	}
	return newMessageContainer(c.ma.Get(mk).Message(), false), nil
}

// func (v *mapElementValue) AccessReadOnly(name string) (Value, error) {
// 	if _, ok := v.ma.NewValue().Interface().(proto.Message); !ok {
// 		return nil, ErrKeyNotFound{Key: name}
// 	}
// 	return newMessageContainer(v.ma.Get(v.key).Message(), false).AccessReadOnly(name)
// }

// func (v *mapElementValue) AccessMutable(name string) (Value, error) {
// 	if _, ok := v.ma.NewValue().Interface().(proto.Message); !ok {
// 		return nil, ErrKeyNotFound{Key: name}
// 	}
// 	return newMessageContainer(v.ma.Get(v.key).Message(), false).AccessMutable(name)
// }

// func (v *scalarValue) AccessReadOnly(name string) (Value, error) {
// 	return nil, ErrFieldNotFound{Field: name}
// }

// func (v *scalarValue) AccessMutable(name string) (Value, error) {
// 	return nil, ErrFieldNotFound{Field: name}
// }
