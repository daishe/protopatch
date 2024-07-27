package protopatch

// func Delete(base protoreflect.Message, path string) error {
// 	err := deleteInMessage(base, path)
// 	if r := recover(); r != nil {
// 		if err, ok := r.(error); ok {
// 			return fmt.Errorf("proto panic recovered: %w", NewErrInPath(path, err))
// 		}
// 		return fmt.Errorf("proto panic recovered: %w", NewErrInPath(path, fmt.Errorf("%v", r)))
// 	}
// 	return err
// }

// func deleteInMessage(base protoreflect.Message, path string) error {
// 	if path == "" {
// 		return ErrDeleteNonKey
// 	}

// 	name, path := Cut(path)
// 	field, err := fieldInMessage(base.Descriptor().Fields(), name)
// 	if err != nil {
// 		return err
// 	}
// 	if field.IsList() {
// 		return NewErrInPath(name, deleteInList(base, field, path))
// 	}
// 	if field.IsMap() {
// 		return NewErrInPath(name, deleteInMap(base, field, path))
// 	}
// 	if field.Kind() == protoreflect.MessageKind {
// 		// TOOD: Oneof check
// 		return NewErrInPath(name, deleteInMessage(base.Mutable(field).Message(), path))
// 	}
// 	return NewErrInPath(name, ErrDeleteNonKey)
// }

// func deleteInList(base protoreflect.Message, listField protoreflect.FieldDescriptor, path string) error {
// 	if path == "" {
// 		return ErrDeleteNonKey
// 	}

// 	name, path := Cut(path)
// 	if name == "*" && path == "" {
// 		if !base.Has(listField) {
// 			return nil
// 		}
// 		base.Mutable(listField).List().Truncate(0)
// 		return nil
// 	}
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
// 			return NewErrInPath(name, deleteInMessage(li.Get(idx).Message(), path))
// 		}
// 		notFound, _ := Cut(path)
// 		return NewErrInPath(name, ErrNotFound{Kind: "field", Value: notFound})
// 	}

// 	len := li.Len()
// 	for i := idx; i < len-1; i++ {
// 		li.Set(i, li.Get(i+1))
// 	}
// 	li.Truncate(len - 1)
// 	return nil

// }

// func deleteInMap(base protoreflect.Message, mapField protoreflect.FieldDescriptor, path string) error {
// 	if path == "" {
// 		return ErrDeleteNonKey
// 	}

// 	name, path := Cut(path)
// 	if name == "*" && path == "" {
// 		if !base.Has(mapField) {
// 			return nil
// 		}
// 		base.Set(mapField, base.NewField(mapField))
// 		return nil
// 	}
// 	if !base.Has(mapField) {
// 		return ErrNotFound{Kind: "field", Value: name}
// 	}
// 	m := base.Mutable(mapField).Map()
// 	key, err := keyInMap(m, mapField.MapKey(), name)
// 	if err != nil {
// 		return err
// 	}
// 	if path != "" {
// 		mapValue := mapField.MapValue()
// 		if mapValue.Kind() == protoreflect.MessageKind {
// 			return NewErrInPath(name, deleteInMessage(m.Get(key).Message(), path))
// 		}
// 		notFound, _ := Cut(path)
// 		return NewErrInPath(name, ErrNotFound{Kind: "field", Value: notFound})
// 	}

// 	m.Delete(key)
// 	return nil
// }
