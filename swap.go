package protopatch

import "google.golang.org/protobuf/proto"

func Swap(base proto.Message, targetPath, replacementPath string, opts ...Option) error {
	return swapWithSetup(base, targetPath, replacementPath, newSetup(opts...))
}

func swapWithSetup(base proto.Message, firstPath, secondPath string, setup *setup) error {
	if firstPath == secondPath { // set value pointed by path to itself
		return setToItself(base, firstPath, setup)
	}

	firstValue, firstSet, err := getCopyAndSetter(base, firstPath, setup)
	if err != nil {
		return err
	}
	secondValue, secondSet, err := getCopyAndSetter(base, secondPath, setup)
	if err != nil {
		return err
	}

	restore := func() {
		_ = firstSet(firstValue)   // ignore errors
		_ = secondSet(secondValue) // ignore errors
	}
	attemptRestore := func() {
		if restore != nil {
			restore()
		}
	}
	defer attemptRestore()

	if secondPath != "" { // avoid setting first value if second refers to the base message
		err = firstSet(secondValue)
		if err != nil {
			return err
		}
	}
	if firstPath != "" { // avoid setting second value if first refers to the base message
		err = secondSet(firstValue)
		if err != nil {
			return err
		}
	}
	restore = nil
	return nil
}

// func Swap(base protoreflect.Message, firstPath, secondPath string) error {
// 	first, err := getValueForSwapInMessage(base, firstPath)
// 	if err != nil {
// 		return err
// 	}
// 	if r := recover(); r != nil {
// 		if err, ok := r.(error); ok {
// 			return fmt.Errorf("proto panic recovered: %w", NewErrInPath(firstPath, err))
// 		}
// 		return fmt.Errorf("proto panic recovered: %w", NewErrInPath(firstPath, fmt.Errorf("%v", r)))
// 	}

// 	if firstPath == secondPath {
// 		return nil
// 	}
// 	second, err := getValueForSwapInMessage(base, secondPath)
// 	if err != nil {
// 		return err
// 	}
// 	if r := recover(); r != nil {
// 		if err, ok := r.(error); ok {
// 			return fmt.Errorf("proto panic recovered: %w", NewErrInPath(secondPath, err))
// 		}
// 		return fmt.Errorf("proto panic recovered: %w", NewErrInPath(secondPath, fmt.Errorf("%v", r)))
// 	}

// 	switch {
// 	case first.listItem != nil && second.messageField != nil:
// 		first, second = second, first
// 	case first.mapItem != nil && second.messageField != nil:
// 		first, second = second, first
// 	case first.mapItem != nil && second.listItem != nil:
// 		first, second = second, first
// 	}

// 	switch {
// 	case first.messageField != nil && second.messageField != nil:
// 		if first.messageField.IsList() != second.messageField.IsList() {
// 			return ErrSwapMissmatchingType{FirstPath: firstPath, SecondPath: secondPath}
// 		}
// 		if first.messageField.IsMap() != second.messageField.IsMap() {
// 			return ErrSwapMissmatchingType{FirstPath: firstPath, SecondPath: secondPath}
// 		}
// 		if !swapIsMatchingType(first.messageField, second.messageField) {
// 			return ErrSwapMissmatchingType{FirstPath: firstPath, SecondPath: secondPath}
// 		}

// 	case first.messageField != nil && second.listItem != nil:
// 		if first.messageField.IsList() {
// 			return ErrSwapMissmatchingType{FirstPath: firstPath, SecondPath: secondPath}
// 		}
// 		if first.messageField.IsMap() {
// 			return ErrSwapMissmatchingType{FirstPath: firstPath, SecondPath: secondPath}
// 		}
// 		if !swapIsMatchingType(first.messageField, second.listItem) {
// 			return ErrSwapMissmatchingType{FirstPath: firstPath, SecondPath: secondPath}
// 		}

// 	case first.messageField != nil && second.mapItem != nil:
// 		if first.messageField.IsList() {
// 			return ErrSwapMissmatchingType{FirstPath: firstPath, SecondPath: secondPath}
// 		}
// 		if first.messageField.IsMap() {
// 			return ErrSwapMissmatchingType{FirstPath: firstPath, SecondPath: secondPath}
// 		}
// 		if !swapIsMatchingType(first.messageField, second.listItem) {
// 			return ErrSwapMissmatchingType{FirstPath: firstPath, SecondPath: secondPath}
// 		}

// 	case first.listItem != nil && second.listItem != nil:
// 		if !swapIsMatchingType(first.listItem, second.listItem) {
// 			return ErrSwapMissmatchingType{FirstPath: firstPath, SecondPath: secondPath}
// 		}

// 	case first.listItem != nil && second.mapItem != nil:
// 		if !swapIsMatchingType(first.listItem, second.mapItem) {
// 			return ErrSwapMissmatchingType{FirstPath: firstPath, SecondPath: secondPath}
// 		}

// 	case first.mapItem != nil && second.mapItem != nil:
// 		if !swapIsMatchingType(first.mapItem, second.mapItem) {
// 			return ErrSwapMissmatchingType{FirstPath: firstPath, SecondPath: secondPath}
// 		}
// 	}

// 	firstVal, secondVal := first.get(), second.get()
// 	first.set(secondVal)
// 	second.set(firstVal)
// 	if r := recover(); r != nil {
// 		if err, ok := r.(error); ok {
// 			return fmt.Errorf("proto panic recovered: %w", err)
// 		}
// 		return fmt.Errorf("proto panic recovered: %w", fmt.Errorf("%v", r))
// 	}
// 	return err
// }

// func swapIsMatchingType(first protoreflect.FieldDescriptor, second protoreflect.FieldDescriptor) bool {
// 	kindMatch := first.Kind() == second.Kind()
// 	enumMatch := first.Enum() == second.Enum()
// 	messageMatch := first.Message() == second.Message()
// 	mapKeyMatch := first.MapKey() == second.MapKey()
// 	mapValueMatch := first.MapValue() == second.MapValue()
// 	return kindMatch && enumMatch && messageMatch && mapKeyMatch && mapValueMatch
// }

// type valueForSwap struct {
// 	messageField protoreflect.FieldDescriptor
// 	listItem     protoreflect.FieldDescriptor
// 	mapItem      protoreflect.FieldDescriptor
// 	get          func() protoreflect.Value
// 	set          func(protoreflect.Value)
// }

// func getValueForSwapInMessage(base protoreflect.Message, path string) (*valueForSwap, error) {
// 	if path == "" {
// 		// m := base.New()
// 		// if err := structPbValueToMessage(m.Interface(), to); err != nil {
// 		// 	return nil, err
// 		// }
// 		// fields := base.Descriptor().Fields()
// 		// for i := 0; i < fields.Len(); i++ {
// 		// 	field := fields.Get(i)
// 		// 	if m.Has(field) {
// 		// 		base.Set(field, m.Get(field))
// 		// 	} else {
// 		// 		base.Clear(field)
// 		// 	}
// 		// }
// 		return nil, nil // TODO: Item by item swap
// 	}

// 	name, path := Cut(path)
// 	field, err := fieldInMessage(base.Descriptor().Fields(), name)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if field.IsList() {
// 		v, err := getValueForSwapInList(base, field, path)
// 		return v, NewErrInPath(name, err)
// 	}
// 	if field.IsMap() {
// 		v, err := getValueForSwapInMap(base, field, path)
// 		return v, NewErrInPath(name, err)
// 	}
// 	if path != "" {
// 		if field.Kind() == protoreflect.MessageKind {
// 			// TODO: Oneof check
// 			v, err := getValueForSwapInMessage(base.Mutable(field).Message(), path)
// 			return v, NewErrInPath(name, err)
// 		}
// 		notFound, _ := Cut(path)
// 		return nil, NewErrInPath(name, ErrNotFound{Kind: "field", Value: notFound})
// 	}
// 	v := &valueForSwap{
// 		messageField: field,
// 		get:          func() protoreflect.Value { return base.Get(field) },
// 		set:          func(to protoreflect.Value) { base.Set(field, to) },
// 	}
// 	return v, nil
// }

// func getValueForSwapInList(base protoreflect.Message, listField protoreflect.FieldDescriptor, path string) (*valueForSwap, error) {
// 	if path == "" {
// 		v := &valueForSwap{
// 			messageField: listField,
// 			get:          func() protoreflect.Value { return base.Get(listField) },
// 			set:          func(to protoreflect.Value) { base.Set(listField, to) },
// 		}
// 		return v, nil
// 	}

// 	name, path := Cut(path)
// 	if !base.Has(listField) {
// 		return nil, ErrNotFound{Kind: "field", Value: name}
// 	}
// 	li := base.Mutable(listField).List()
// 	idx, err := indexInList(li, name)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if path != "" {
// 		if listField.Kind() == protoreflect.MessageKind {
// 			v, err := getValueForSwapInMessage(li.Get(idx).Message(), path)
// 			return v, NewErrInPath(name, err)
// 		}
// 		notFound, _ := Cut(path)
// 		return nil, NewErrInPath(name, ErrNotFound{Kind: "field", Value: notFound})
// 	}
// 	v := &valueForSwap{
// 		listItem: listField,
// 		get:      func() protoreflect.Value { return li.Get(idx) },
// 		set:      func(to protoreflect.Value) { li.Set(idx, to) },
// 	}
// 	return v, nil
// }

// func getValueForSwapInMap(base protoreflect.Message, mapField protoreflect.FieldDescriptor, path string) (*valueForSwap, error) {
// 	if path == "" {
// 		v := &valueForSwap{
// 			messageField: mapField,
// 			get:          func() protoreflect.Value { return base.Get(mapField) },
// 			set:          func(to protoreflect.Value) { base.Set(mapField, to) },
// 		}
// 		return v, nil
// 	}

// 	name, path := Cut(path)
// 	if !base.Has(mapField) {
// 		return nil, ErrNotFound{Kind: "key", Value: name}
// 	}
// 	m := base.Mutable(mapField).Map()
// 	key, err := keyInMap(m, mapField.MapKey(), name)
// 	if err != nil {
// 		return nil, err
// 	}
// 	mapValue := mapField.MapValue()
// 	if path != "" {
// 		if mapValue.Kind() == protoreflect.MessageKind {
// 			v, err := getValueForSwapInMessage(m.Get(key).Message(), path)
// 			return v, NewErrInPath(name, err)
// 		}
// 		notFound, _ := Cut(path)
// 		return nil, NewErrInPath(name, ErrNotFound{Kind: "field", Value: notFound})
// 	}
// 	v := &valueForSwap{
// 		mapItem: mapValue,
// 		get:     func() protoreflect.Value { return m.Get(key) },
// 		set:     func(to protoreflect.Value) { m.Set(key, to) },
// 	}
// 	return v, nil
// }
