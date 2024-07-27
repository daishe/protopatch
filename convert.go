package protopatch

import "reflect"

// func (v *messageContainer) Convert(to any) (any, error) {
// 	if reflect.TypeOf(v.msg.Interface()) == reflect.TypeOf(to) {
// 		return v.msg.Interface(), nil
// 	}

// 	m, mOk := to.(proto.Message)
// 	pr, prOk := to.(protoreflect.Message)
// 	if !mOk && prOk {
// 		return nil, errors.New("cannot convert")
// 	}
// 	if mOk {
// 		pr = m.ProtoReflect()
// 	}
// 	if v.msg.Descriptor() != pr.Descriptor() {
// 		return nil, errors.New("cannot convert")
// 	}
// 	return v.msg.Interface(), nil
// }

// func (v *messageFieldValue) Convert(to any) (any, error) {
// 	if reflect.TypeOf(v.msg.Interface()) == reflect.TypeOf(to) {
// 		return v.msg.Interface(), nil
// 	}

// 	m, mOk := to.(proto.Message)
// 	pr, prOk := to.(protoreflect.Message)
// 	if !mOk && prOk {
// 		return nil, errors.New("cannot convert")
// 	}
// 	if mOk {
// 		pr = m.ProtoReflect()
// 	}
// 	if v.msg.Descriptor() != pr.Descriptor() {
// 		return nil, errors.New("cannot convert")
// 	}
// 	return v.msg.Interface(), nil
// }

// func (v *listContainer) Convert(to any) (any, error) {
// 	if i := v.Interface(); reflect.TypeOf(i) == reflect.TypeOf(to) {
// 		return i, nil
// 	}
// 	return nil, errors.New("cannot convert")
// }

// func (v *listElementValue) Convert(to any) (any, error) {
// 	if i := v.Interface(); reflect.TypeOf(i) == reflect.TypeOf(to) {
// 		return i, nil
// 	}
// 	return nil, errors.New("cannot convert")
// }

// func (v *mapContainer) Convert(to any) (any, error) {
// 	if i := v.Interface(); reflect.TypeOf(i) == reflect.TypeOf(to) {
// 		return i, nil
// 	}
// 	return nil, errors.New("cannot convert")
// }

// func (v *mapElementValue) Convert(to any) (any, error) {
// 	if i := v.Interface(); reflect.TypeOf(i) == reflect.TypeOf(to) {
// 		return i, nil
// 	}
// 	return nil, errors.New("cannot convert")
// }

// func (v *scalarValue) Convert(to any) (any, error) {
// 	if i := v.Interface(); reflect.TypeOf(i) == reflect.TypeOf(to) {
// 		return i, nil
// 	}
// 	return nil, errors.New("cannot convert")
// }

// IdentityConverter converts the provided value to itself ensuring that provided types match. It returns ErrNoConversionDefined error for mismatched types.
func IdentityConverter(to, from any) (any, error) {
	toVal, fromVal := reflect.ValueOf(to), reflect.ValueOf(from)

	// proto.Message, scalars, go types
	if areValueTypesMatch(toVal, fromVal) {
		return from, nil
	}

	// List to List
	toList, toListOk := to.(List)
	fromList, fromListOk := from.(List)
	if toListOk && fromListOk && areProtoFieldMatch(toList.ParentFieldDescriptor(), fromList.ParentFieldDescriptor()) {
		return from, nil
	}
	// slice to List
	if toListOk && isValueTypeMatchesProtoField(toList.ParentFieldDescriptor(), fromVal) {
		return from, nil
	}
	// List to slice
	if fromListOk && isValueTypeMatchesProtoField(fromList.ParentFieldDescriptor(), toVal) {
		return from, nil
	}

	// Map to Map
	toMap, toMapOk := to.(Map)
	fromMap, fromMapOk := from.(Map)
	if toMapOk && fromMapOk && areProtoFieldMatch(toMap.ParentFieldDescriptor(), fromMap.ParentFieldDescriptor()) {
		return from, nil
	}
	// map to Map
	if toMapOk && isValueTypeMatchesProtoField(toMap.ParentFieldDescriptor(), fromVal) {
		return from, nil
	}
	// Map to map
	if fromMapOk && isValueTypeMatchesProtoField(fromMap.ParentFieldDescriptor(), toVal) {
		return from, nil
	}

	return nil, ErrNoConversionDefined
}
