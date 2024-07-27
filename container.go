package protopatch

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// type ContainerKind uint32

// const (
// 	KindOther ContainerKind = iota
// 	KindMessage
// 	KindList
// 	KindMap
// )

// // ContainerDescriptor contains basic information about the container.
// type ContainerDescriptor struct {
// 	kind    ContainerKind
// 	message protoreflect.MessageDescriptor
// 	field   protoreflect.FieldDescriptor
// }

// // Kind returns kind of the container.
// func (cd ContainerDescriptor) Kind() ContainerKind {
// 	return cd.kind
// }

// // Message returns protoreflect.MessageDescriptor associated with:
// // - the underlying message for KindMessage;
// // - the message congaing list field for KindList;
// // - the message containing map field for KindMap.
// // For KindOther value returned is unspecified and can be nil.
// func (cd ContainerDescriptor) Message() protoreflect.MessageDescriptor {
// 	return cd.message
// }

// // Field returns protoreflect.FieldDescriptor associated with:
// // - the list field for KindList;
// // - the map field for KindMap.
// // For KindMessage it returns nil. For KindOther value returned is unspecified and can be nil.
// func (cd ContainerDescriptor) Field() protoreflect.FieldDescriptor {
// 	return cd.field
// }

// Container is a wrapper of a composite protocol buffer type (message, list or map) allowing easy traversal and operations.
type Container interface {
	// // Descriptor returns descriptor associated with the container.
	// Descriptor() ContainerDescriptor

	// IsReadOnly reports wether underlying container value is read only.
	IsReadOnly() bool

	// Self returns underlying container value. For messages it returns its proto.Message value. For lists and maps it returns List and Map interfaces accordingly.
	Self() any

	// Get returns value associated with the given field / index / key. It returns error if the field / index / key is not found. For scalar types and messages it returns its value. For lists and maps it returns List and Map interfaces accordingly.
	Get(string) (any, error)

	// GetCopy returns a copy of value associated with the given field / index / key. It returns error if the field / index / key is not found. For scalar types and messages it returns its value. For lists and maps it returns List and Map interfaces accordingly.
	GetCopy(string) (any, error)

	// GetNew returns a zero value of the type of value associated with the given field / index / key. It returns error if the field is not found or if index / key is malformed. For scalar types and messages it returns its zero value. For lists and maps it returns List and Map interfaces accordingly without any elements.
	GetNew(string) (any, error)

	// Mutable is a mutable variant of Get method - it returns value associated with the given field / index / key. It returns error if the container is read-only or the field / index / key is not found. For scalar types and messages it returns its value. For lists and maps it returns List and Map interfaces accordingly.
	Mutable(string) (any, error)

	// Access descends into the given field / index / key and returns a new container for that value. It returns error if the field / index / key is not found or is not associated with a composite type.
	Access(string) (Container, error)

	// AccessMutable is a mutable variant of Access method - it descends into the given field / index / key and returns a new container for that value. It returns error if the container is read-only, if the field / index / key is not found or is not associated with a composite type.
	AccessMutable(string) (Container, error)

	// Set stores the provided value under the associated field / index / key. It returns error if the container is read-only, if the field / index / key is not found or if the value has wrong type. When the provided value is nil, it clears the associated field / index / key.
	Set(string, any) error

	// Append adds the provided value to the end of a list. It returns error if the container is read-only, if the container is not a list or if the value has wrong type.
	Append(any) error

	// Insert adds the provided value at specified position to a list or map. It returns error if the container is read-only, if the container is not a list or map container, if the index is not found or if the value has wrong type.
	Insert(string, any) error
}

type messageContainer struct {
	msg protoreflect.Message
	ro  bool
}

func MessageContainer(m proto.Message) Container {
	pr := m.ProtoReflect()
	return newMessageContainer(pr, !pr.IsValid())
}

func newMessageContainer(msg protoreflect.Message, ro bool) *messageContainer {
	return &messageContainer{msg: msg, ro: ro}
}

// func (c *messageContainer) Descriptor() ContainerDescriptor {
// 	return ContainerDescriptor{kind: KindMessage, message: c.msg.Descriptor()}
// }

func (c *messageContainer) IsReadOnly() bool {
	return c.ro
}

// type messageFieldValue struct {
// 	msg   protoreflect.Message
// 	field protoreflect.FieldDescriptor
// 	ro    bool
// }

// func newMessageFieldValue(msg protoreflect.Message, field protoreflect.FieldDescriptor, ro bool) *messageFieldValue {
// 	return &messageFieldValue{msg: msg, field: field, ro: ro}
// }

// func (v *messageFieldValue) Interface() any {
// 	return v.msg.Get(v.field)
// }

// func (v *messageFieldValue) IsReadOnly() bool {
// 	return !v.msg.Has(v.field)
// }

type listContainer struct {
	parent      protoreflect.Message
	parentField protoreflect.FieldDescriptor
	li          protoreflect.List
	ro          bool
}

func newListContainer(m protoreflect.Message, f protoreflect.FieldDescriptor, li protoreflect.List, ro bool) *listContainer {
	return &listContainer{parent: m, parentField: f, li: li, ro: ro}
}

// func (c *listContainer) Descriptor() ContainerDescriptor {
// 	return ContainerDescriptor{kind: KindList, message: c.parent.Descriptor(), field: c.parentField}
// }

func (c *listContainer) IsReadOnly() bool {
	return c.ro
}

// type listElementValue struct {
// 	li  protoreflect.List
// 	idx int
// }

// func newListElementValue(li protoreflect.List, idx int) *listElementValue {
// 	return &listElementValue{li: li, idx: idx}
// }

// func (v *listElementValue) Interface() any {
// 	return v.li.Get(v.idx)
// }

// func (v *listElementValue) IsReadOnly() bool {
// 	return false
// }

type mapContainer struct {
	parent      protoreflect.Message
	parentField protoreflect.FieldDescriptor
	ma          protoreflect.Map
	ro          bool
}

func newMapContainer(m protoreflect.Message, f protoreflect.FieldDescriptor, ma protoreflect.Map, ro bool) *mapContainer {
	return &mapContainer{parent: m, parentField: f, ma: ma, ro: ro}
}

// func (c *mapContainer) Descriptor() ContainerDescriptor {
// 	return ContainerDescriptor{kind: KindMap, message: c.parent.Descriptor(), field: c.parentField}
// }

func (c *mapContainer) IsReadOnly() bool {
	return c.ro
}

// type mapElementValue struct {
// 	ma  protoreflect.Map
// 	key protoreflect.MapKey
// }

// func newMapElementValue(ma protoreflect.Map, key protoreflect.MapKey) *mapElementValue {
// 	return &mapElementValue{ma: ma, key: key}
// }

// func (v *mapElementValue) Interface() any {
// 	return v.ma.Get(v.key)
// }

// func (v *mapElementValue) IsReadOnly() bool {
// 	return false
// }

// type scalarValue struct {
// 	val any
// }

// func newScalarValue(val any) *scalarValue {
// 	return &scalarValue{val: val}
// }

// func (v *scalarValue) Interface() any {
// 	return v.val
// }

// func (v *scalarValue) IsReadOnly() bool {
// 	return true
// }
