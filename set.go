package protopatch

import (
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// func Set(base protoreflect.Message, path string, to *structpb.Value) error {
// 	err := setInMessage(base, path, to)
// 	if r := recover(); r != nil {
// 		if err, ok := r.(error); ok {
// 			return fmt.Errorf("proto panic recovered: %w", NewErrInPath(path, err))
// 		}
// 		return fmt.Errorf("proto panic recovered: %w", NewErrInPath(path, fmt.Errorf("%v", r)))
// 	}
// 	return err
// }

// func setInMessage(base protoreflect.Message, path string, to *structpb.Value) error {
// 	if path == "" {
// 		m := base.New()
// 		if err := structPbValueToMessage(m.Interface(), to); err != nil {
// 			return err
// 		}
// 		fields := base.Descriptor().Fields()
// 		for i := 0; i < fields.Len(); i++ {
// 			field := fields.Get(i)
// 			if m.Has(field) {
// 				base.Set(field, m.Get(field))
// 			} else {
// 				base.Clear(field)
// 			}
// 		}
// 		return nil
// 	}

// 	name, path := Cut(path)
// 	field, err := fieldInMessage(base.Descriptor().Fields(), name)
// 	if err != nil {
// 		return err
// 	}
// 	if field.IsList() {
// 		return NewErrInPath(name, setInList(base, field, path, to))
// 	}
// 	if field.IsMap() {
// 		return NewErrInPath(name, setInMap(base, field, path, to))
// 	}
// 	if field.Kind() == protoreflect.MessageKind {
// 		if path == "" {
// 			v := base.NewField(field)
// 			if err := structPbValueToMessage(v.Message().Interface(), to); err != nil {
// 				return NewErrInPath(name, err)
// 			}
// 			base.Set(field, v)
// 			return nil
// 		}
// 		// TODO: Oneof check
// 		return NewErrInPath(name, setInMessage(base.Mutable(field).Message(), path, to))
// 	}
// 	if path != "" {
// 		notFound, _ := Cut(path)
// 		return NewErrInPath(name, ErrFieldNotFound{Field: notFound})
// 	}
// 	scalar, err := structPbValueToScalar(field, to)
// 	if err != nil {
// 		return NewErrInPath(name, err)
// 	}
// 	base.Set(field, scalar)
// 	return nil
// }

// func setInList(base protoreflect.Message, listField protoreflect.FieldDescriptor, path string, to *structpb.Value) error {
// 	if path == "" {
// 		return structPbValueToList(base, listField, to)
// 	}

// 	name, path := Cut(path)
// 	if !base.Has(listField) {
// 		return ErrFieldNotFound{Field: name}
// 	}
// 	li := base.Mutable(listField).List()
// 	idx, err := indexInList(li, name)
// 	if err != nil {
// 		return err
// 	}
// 	if listField.Kind() == protoreflect.MessageKind {
// 		return NewErrInPath(name, setInMessage(li.Get(idx).Message(), path, to))
// 	}
// 	if path != "" {
// 		notFound, _ := Cut(path)
// 		return NewErrInPath(name, ErrFieldNotFound{Field: notFound})
// 	}
// 	scalar, err := structPbValueToScalar(listField, to)
// 	if err != nil {
// 		return NewErrInPath(name, err)
// 	}
// 	li.Set(idx, scalar)
// 	return nil
// }

// func setInMap(base protoreflect.Message, mapField protoreflect.FieldDescriptor, path string, to *structpb.Value) error {
// 	mapValue := mapField.MapValue()
// 	if path == "" {
// 		return structPbValueToMap(base, mapField, to)
// 	}

// 	name, path := Cut(path)
// 	if !base.Has(mapField) {
// 		return ErrKeyNotFound{Key: name}
// 	}
// 	m := base.Mutable(mapField).Map()
// 	key, err := keyInMap(m, mapField.MapKey(), name)
// 	if err != nil {
// 		return err
// 	}
// 	if mapValue.Kind() == protoreflect.MessageKind {
// 		return NewErrInPath(name, setInMessage(m.Get(key).Message(), path, to))
// 	}
// 	if path != "" {
// 		notFound, _ := Cut(path)
// 		return NewErrInPath(name, ErrFieldNotFound{Field: notFound})
// 	}
// 	scalar, err := structPbValueToScalar(mapValue, to)
// 	if err != nil {
// 		return NewErrInPath(name, err)
// 	}
// 	m.Set(key, scalar)
// 	return nil
// }

func Set(base proto.Message, path string, to any, opts ...Option) error {
	return setWithSetup(base, path, to, newSetup(opts...))
}

func setWithSetup(base proto.Message, path string, to any, setup *setup) error {
	if to == nil {
		return clearWithSetup(base, path, setup)
	}
	if path == "" { // special case - an empty path; set of the base message
		return setSelf(base, to, setup)
	}

	c := MessageContainer(base)
	p := Path(path)

	if last := p.Last(); !last.IsFirst() { // path has more than 1 element
		a, err := access(c, last.PrecedingPath(), setup)
		if err != nil {
			return err
		}
		ref, err := a.GetNew(last.Value())
		if err != nil {
			return NewErrInPath(string(last.PrecedingPath()), err)
		}
		conv, err := convert(ref, to, setup)
		if err != nil {
			return NewErrInPath(string(last.PrecedingPathWithCurrentSegment()), err)
		}
		err = a.Set(last.Value(), conv)
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
	ref, err := a.GetNew(path)
	if err != nil {
		return err
	}
	conv, err := convert(ref, to, setup)
	if err != nil {
		return NewErrInPath(path, err)
	}
	err = c.Set(path, conv)
	if err != nil {
		return err
	}
	return nil
}

func setSelf(base proto.Message, to any, setup *setup) error {
	if to == nil {
		return clearSelf(base, setup)
	}
	c := MessageContainer(base).(*messageContainer)
	conv, err := convert(c.Self(), to, setup)
	if err != nil {
		return err
	}
	return c.setSelf(conv)
}

func setToItself(base proto.Message, path string, setup *setup) error {
	c := MessageContainer(base)
	p := Path(path)
	if last := p.Last(); !last.IsFirst() { // path has more than 1 element
		a, err := access(c, p, setup)
		if err != nil {
			return err
		}
		v, err := a.Get(last.Value())
		if err != nil {
			return NewErrInPath(string(last.PrecedingPath()), err)
		}
		err = a.Set(last.Value(), v)
		if err != nil {
			return NewErrInPath(string(last.PrecedingPath()), err)
		}
		return nil
	}
	// path has only 1 element or points to the base message
	a, err := transformContainer(c, setup)
	if err != nil {
		return err
	}
	if p == "" { // path points to the base message
		return nil
	}
	v, err := a.Get(path)
	if err != nil {
		return err
	}
	err = a.Set(path, v)
	if err != nil {
		return err
	}
	return nil
}

func getCopyAndSetter(base proto.Message, path string, setup *setup) (any, func(any) error, error) {
	if path == "" { // special case - an empty path; set of the base message
		original := proto.Clone(base)
		setFn := func(to any) error {
			return setSelf(base, to, setup)
		}
		return original, setFn, nil
	}

	c := MessageContainer(base)
	p := Path(path)

	if last := p.Last(); !last.IsFirst() { // path has more than 1 element
		a, err := access(c, last.PrecedingPath(), setup)
		if err != nil {
			return nil, nil, err
		}
		original, err := a.GetCopy(last.Value())
		if err != nil {
			return nil, nil, NewErrInPath(string(last.PrecedingPath()), err)
		}
		setFn := func(to any) error {
			if to != nil {
				ref, err := a.GetNew(last.Value())
				if err != nil {
					return NewErrInPath(string(last.PrecedingPath()), err)
				}
				to, err = convert(ref, to, setup)
				if err != nil {
					return NewErrInPath(string(last.PrecedingPathWithCurrentSegment()), err)
				}
			}
			err = a.Set(last.Value(), to)
			if err != nil {
				return NewErrInPath(string(last.PrecedingPath()), err)
			}
			return nil
		}
		return original, setFn, nil
	}

	// path has only 1 element
	a, err := transformContainer(c, setup)
	if err != nil {
		return nil, nil, err
	}
	original, err := a.GetCopy(path)
	if err != nil {
		return nil, nil, err
	}
	setFn := func(to any) error {
		if to != nil {
			ref, err := a.GetNew(path)
			if err != nil {
				return err
			}
			to, err = convert(ref, to, setup)
			if err != nil {
				return NewErrInPath(path, err)
			}
		}
		err = c.Set(path, to)
		if err != nil {
			return err
		}
		return nil
	}
	return original, setFn, nil
}

func (c *messageContainer) setSelf(to any) error {
	if to == nil {
		return c.clearSelf()
	}
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	pr := asProtoreflectMessage(to)
	if pr == nil {
		return newSetFailure(ErrMismatchingType)
	}
	if c.msg.Descriptor() != pr.Descriptor() {
		return newSetFailure(ErrMismatchingType)
	}
	c.copyFields(pr)
	return nil
}

func (c *messageContainer) copyFields(from protoreflect.Message) {
	fields := c.msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if !from.Has(field) {
			c.msg.Clear(field)
		} else {
			c.msg.Set(field, from.Get(field))
		}
	}
}

func (c *messageContainer) Set(key string, to any) error {
	if to == nil {
		return c.clear(key)
	}
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	field, err := fieldInMessage(c.msg.Descriptor().Fields(), key)
	if err != nil {
		return err
	}
	if field.IsList() {
		return NewErrInPath(key, c.setList(field, to))
	}
	if field.IsMap() {
		return NewErrInPath(key, c.setMap(field, to))
	}
	if field.Kind() == protoreflect.MessageKind {
		return NewErrInPath(key, c.setMessage(field, to))
	}
	if !isTypeMatchesProtoScalarKind(field.Kind(), reflect.TypeOf(to)) {
		return NewErrInPath(key, newSetFailure(ErrMismatchingType))
	}
	c.msg.Set(field, protoreflect.ValueOf(to))
	return nil
}

func (c *messageContainer) setMessage(field protoreflect.FieldDescriptor, to any) error {
	pr := asProtoreflectMessage(to)
	if pr == nil {
		return newSetFailure(ErrMismatchingType)
	}
	if field.Message() != pr.Descriptor() {
		return newSetFailure(ErrMismatchingType)
	}
	c.msg.Set(field, protoreflect.ValueOfMessage(pr))
	return nil
}

func (c *messageContainer) setList(field protoreflect.FieldDescriptor, to any) error {
	if li, ok := to.(List); ok {
		to = li.AsGoSlice()
	}
	v := reflect.ValueOf(to)
	if !isValueTypeMatchesProtoField(field, v) {
		return newSetFailure(ErrMismatchingType)
	}
	liVal := c.msg.NewField(field)
	li := liVal.List()
	if field.Kind() == protoreflect.MessageKind {
		for _, i := range v.Seq2() {
			li.Append(protoreflect.ValueOfMessage(asProtoreflectMessage(i.Interface())))
		}
		c.msg.Set(field, liVal)
		return nil
	}
	for _, i := range v.Seq2() {
		li.Append(protoreflect.ValueOf(i.Interface()))
	}
	c.msg.Set(field, liVal)
	return nil
}

func (c *messageContainer) setMap(field protoreflect.FieldDescriptor, to any) error {
	if ma, ok := to.(Map); ok {
		to = ma.AsGoMap()
	}
	v := reflect.ValueOf(to)
	if !isValueTypeMatchesProtoField(field, v) {
		return newSetFailure(ErrMismatchingType)
	}
	maVal := c.msg.NewField(field)
	ma := maVal.Map()
	if field.MapValue().Kind() == protoreflect.MessageKind {
		for k, el := range v.Seq2() {
			ma.Set(protoreflect.ValueOf(k.Interface()).MapKey(), protoreflect.ValueOfMessage(asProtoreflectMessage(el.Interface())))
		}
		c.msg.Set(field, maVal)
		return nil
	}
	for k, el := range v.Seq2() {
		ma.Set(protoreflect.ValueOf(k.Interface()).MapKey(), protoreflect.ValueOf(el.Interface()))
	}
	c.msg.Set(field, maVal)
	return nil
}

// func (v *messageFieldValue) Set(to Value) error {
// 	if v.ro {
// 		return ErrMutationOfReadOnlyValue
// 	}
// 	i := to.Interface()
// 	iType := reflect.TypeOf(i)
// 	if v.field.Kind() == protoreflect.MessageKind {
// 		m, mOk := i.(proto.Message)
// 		pm, pmOk := i.(protoreflect.Message)
// 		if !mOk && !pmOk {
// 			return errors.New("cannot set: not a message") // TODO: Better error message and type
// 		}
// 		if mOk {
// 			pm = m.ProtoReflect()
// 		}
// 		if v.msg.Descriptor() != pm.Descriptor() {
// 			return errors.New("cannot set: mismatching message descriptor") // TODO: Better error message and type
// 		}
// 		v.msg.Set(v.field, protoreflect.ValueOfMessage(pm))
// 		return nil
// 	}
// 	if reflect.TypeOf(v.Interface()) != iType {
// 		if iType.Kind() == reflect.String {
// 			return v.setFromString(i.(string))
// 		}
// 		return errors.New("cannot set: invalid type") // TODO: Better error message and type
// 	}
// 	v.msg.Set(v.field, protoreflect.ValueOf(i))
// 	return nil
// }

// func (v *messageFieldValue) setFromString(to string) error {
// 	switch v.field.Kind() {
// 	case protoreflect.EnumKind:
// 		val := v.field.Enum().Values().ByName(protoreflect.Name(to))
// 		if val == nil {
// 			return errors.New("cannot set: invalid enum constant") // TODO: Better error message and type
// 		}
// 		v.msg.Set(v.field, protoreflect.ValueOfEnum(val.Number()))
// 		return nil

// 	case protoreflect.BoolKind:
// 		b, err := strconv.ParseBool(to)
// 		if err != nil {
// 			return errors.New("cannot set: invalid int32 value") // TODO: Better error message and type
// 		}
// 		v.msg.Set(v.field, protoreflect.ValueOfBool(b))
// 		return nil

// 	case protoreflect.Int32Kind, protoreflect.Sfixed32Kind:
// 		i, err := strconv.ParseInt(to, 0, 32)
// 		if err != nil {
// 			return errors.New("cannot set: invalid int32 value") // TODO: Better error message and type
// 		}
// 		v.msg.Set(v.field, protoreflect.ValueOfInt32(int32(i)))
// 		return nil

// 	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
// 		u, err := strconv.ParseUint(to, 0, 32)
// 		if err != nil {
// 			return errors.New("cannot set: invalid uint32 value") // TODO: Better error message and type
// 		}
// 		v.msg.Set(v.field, protoreflect.ValueOfUint32(uint32(u)))
// 		return nil

// 	case protoreflect.Int64Kind, protoreflect.Sfixed64Kind:
// 		i, err := strconv.ParseInt(to, 0, 64)
// 		if err != nil {
// 			return errors.New("cannot set: invalid int64 value") // TODO: Better error message and type
// 		}
// 		v.msg.Set(v.field, protoreflect.ValueOfInt64(i))
// 		return nil

// 	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
// 		u, err := strconv.ParseUint(to, 0, 64)
// 		if err != nil {
// 			return errors.New("cannot set: invalid uint64 value") // TODO: Better error message and type
// 		}
// 		v.msg.Set(v.field, protoreflect.ValueOfUint64(u))
// 		return nil

// 	case protoreflect.FloatKind:
// 		f, err := strconv.ParseFloat(to, 32)
// 		if err != nil {
// 			return errors.New("cannot set: invalid float value") // TODO: Better error message and type
// 		}
// 		v.msg.Set(v.field, protoreflect.ValueOfFloat32(float32(f)))
// 		return nil

// 	case protoreflect.DoubleKind:
// 		f, err := strconv.ParseFloat(to, 64)
// 		if err != nil {
// 			return errors.New("cannot set: invalid float value") // TODO: Better error message and type
// 		}
// 		v.msg.Set(v.field, protoreflect.ValueOfFloat64(f))
// 		return nil

// 	case protoreflect.BytesKind:
// 		v.msg.Set(v.field, protoreflect.ValueOfBytes([]byte(to)))
// 		return nil

// 	}

// 	return errors.New("cannot set: invalid type") // TODO: Better error message and type
// }

// func (c *listContainer) setSelf(to any) error {
// 	if c.ro {
// 		return ErrMutationOfReadOnlyValue
// 	}
// 	v := reflect.ValueOf(to)
// 	if c.parentField.Kind() == protoreflect.MessageKind {
// 		m, mOk := reflect.Zero(v.Type().Elem()).Interface().(proto.Message)
// 		if !mOk {
// 			return errors.New("cannot set: not a message list") // TODO: Better error message and type
// 		}
// 		pm := m.ProtoReflect()
// 		if c.parentField.Message() != pm.Descriptor() {
// 			return errors.New("cannot set: mismatching message descriptor in list") // TODO: Better error message and type
// 		}
// 	} else if reflect.TypeOf(protoreflect.ValueOfList(c.li).Interface()) != v.Type() {
// 		return errors.New("cannot set: invalid type") // TODO: Better error message and type
// 	}
// 	if c.li.Len() > v.Len() {
// 		c.li.Truncate(v.Len())
// 	}
// 	for i := 0; i < min(c.li.Len(), v.Len()); i++ {
// 		c.li.Set(i, protoreflect.ValueOf(v.Index(i).Interface()))
// 	}
// 	for i := c.li.Len(); i < v.Len(); i++ {
// 		c.li.Append(protoreflect.ValueOf(v.Index(i).Interface()))
// 	}
// 	return nil
// }

func (c *listContainer) Set(key string, to any) error {
	if to == nil {
		return c.clear(key)
	}
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	idx, err := indexInList(c.li, key)
	if err != nil {
		return err
	}
	if c.parentField.Kind() == protoreflect.MessageKind {
		pr := asProtoreflectMessage(to)
		if pr == nil {
			return NewErrInPath(key, newSetFailure(ErrMismatchingType))
		}
		if c.parentField.Message() != pr.Descriptor() {
			return NewErrInPath(key, newSetFailure(ErrMismatchingType))
		}
		c.li.Set(idx, protoreflect.ValueOfMessage(pr))
		return nil
	}
	if reflect.TypeOf(c.li.Get(idx).Interface()) != reflect.TypeOf(to) {
		return NewErrInPath(key, newSetFailure(ErrMismatchingType))
	}
	c.li.Set(idx, protoreflect.ValueOf(to))
	return nil
}

// func (v *listElementValue) Set(to Value) error {
// 	vInt, toInt := v.Interface(), to.Interface()
// 	if vMsg, ok := vInt.(proto.Message); ok {
// 		toMsg, toMsgOk := toInt.(proto.Message)
// 		toPMsg, toPMsgOk := toInt.(protoreflect.Message)
// 		if !toMsgOk && !toPMsgOk {
// 			return errors.New("cannot set: not a message") // TODO: Better error message and type
// 		}
// 		if toMsgOk {
// 			toPMsg = toMsg.ProtoReflect()
// 		}
// 		if vMsg.ProtoReflect().Descriptor() != toPMsg.Descriptor() {
// 			return errors.New("cannot set: mismatching message descriptor") // TODO: Better error message and type
// 		}
// 		v.li.Set(v.idx, protoreflect.ValueOfMessage(toPMsg))
// 		return nil
// 	}
// 	if vType, toType := reflect.TypeOf(vInt), reflect.TypeOf(toInt); vType != toType {
// 		if toType.Kind() == reflect.String {
// 			x, err := convertStringToGoType(toInt.(string), vType.Kind())
// 			if err != nil {
// 				return err
// 			}
// 			v.li.Set(v.idx, protoreflect.ValueOf(x))
// 			return nil
// 		}
// 		return errors.New("cannot set: invalid type") // TODO: Better error message and type
// 	}
// 	v.li.Set(v.idx, protoreflect.ValueOf(toInt))
// 	return nil
// }

// func (c *mapContainer) setSelf(to any) error {
// 	if c.ro {
// 		return ErrMutationOfReadOnlyValue
// 	}
// 	v := reflect.ValueOf(to)
// 	if c.parentField.MapValue().Kind() == protoreflect.MessageKind {
// 		vType := v.Type()
// 		if vType.Kind() != reflect.Map {
// 			return errors.New("cannot set: not a message map") // TODO: Better error message and type
// 		}
// 		m, mOk := reflect.Zero(vType.Elem()).Interface().(proto.Message)
// 		if !mOk {
// 			return errors.New("cannot set: not a message map") // TODO: Better error message and type
// 		}
// 		pm := m.ProtoReflect()
// 		if c.parentField.Message() != pm.Descriptor() {
// 			return errors.New("cannot set: mismatching message descriptor in map") // TODO: Better error message and type
// 		}
// 		if vType.Key() != reflect.TypeOf(protoreflect.ValueOfMap(c.ma).Interface()).Key() {
// 			return errors.New("cannot set: bad key type") // TODO: Better error message and type
// 		}
// 	} else if reflect.TypeOf(protoreflect.ValueOfMap(c.ma).Interface()) != v.Type() {
// 		return errors.New("cannot set: invalid type") // TODO: Better error message and type
// 	}
// 	c.ma.Range(func(mk protoreflect.MapKey, _ protoreflect.Value) bool {
// 		c.ma.Clear(mk)
// 		return true
// 	})
// 	for it := v.MapRange(); it.Next(); {
// 		c.ma.Set(protoreflect.ValueOf(it.Key().Interface()).MapKey(), protoreflect.ValueOf(it.Value().Interface()))
// 	}
// 	return nil
// }

func (c *mapContainer) Set(key string, to any) error {
	if to == nil {
		return c.clear(key)
	}
	if c.ro {
		return ErrMutationOfReadOnlyValue
	}
	mk, err := keyInMap(c.ma, c.parentField.MapKey(), key)
	if err != nil {
		newMk, parseErr := parseMapKey(c.parentField.MapKey(), key) // check it it is a new map key insertion
		if parseErr != nil {
			return err // return original error
		}
		mk = newMk
	}
	if c.parentField.MapValue().Kind() == protoreflect.MessageKind {
		pr := asProtoreflectMessage(to)
		if pr == nil {
			return NewErrInPath(key, newSetFailure(ErrMismatchingType))
		}
		if c.parentField.MapValue().Message() != pr.Descriptor() {
			return NewErrInPath(key, newSetFailure(ErrMismatchingType))
		}
		c.setCheckedValue(mk, protoreflect.ValueOfMessage(pr))
		return nil
	}
	ref := c.ma.Get(mk).Interface()
	if !c.ma.Has(mk) {
		ref = c.ma.NewValue().Interface()
	}
	if reflect.TypeOf(ref) != reflect.TypeOf(to) {
		return NewErrInPath(key, newSetFailure(ErrMismatchingType))
	}
	c.setCheckedValue(mk, protoreflect.ValueOf(to))
	return nil
}

func (c *mapContainer) setCheckedValue(mk protoreflect.MapKey, to protoreflect.Value) {
	if !c.ma.IsValid() {
		c.ma = c.parent.Mutable(c.parentField).Map()
	}
	c.ma.Set(mk, to)
}
