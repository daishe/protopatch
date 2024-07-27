package patchstructpb

import (
	"github.com/daishe/protopatch"
	"github.com/daishe/protopatch/internal/protoops"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

func ValueContainerTransformer(opts ...Option) protopatch.ContainerTransformer {
	setup := newSetup(opts...)
	return protopatch.ContainerTransformerFunc(func(container protopatch.Container) (protopatch.Container, error) {
		return containerTransform(container, setup)
	})
}

func ValueContainerTransform(container protopatch.Container, opts ...Option) (protopatch.Container, error) {
	return containerTransform(container, newSetup(opts...))
}

func containerTransform(container protopatch.Container, setup *setup) (protopatch.Container, error) {
	switch v := container.Self().(type) {
	case *structpb.Struct:
		if container.IsReadOnly() {
			v = nil
		}
		return &structValueContainer{st: v}, nil
	case *structpb.ListValue:
		if container.IsReadOnly() {
			v = nil
		}
		return &listValueContainer{li: v}, nil
	case *structpb.Value:
		if container.IsReadOnly() {
			v = nil
		}
		return &valueContainer{v: v}, nil
	}
	// TODO: Support container transformation when descriptor matches any structpb message to support dynamic messages.
	return nil, protopatch.ErrNoContainerTransformationDefined
}

type structValueContainer struct {
	st *structpb.Struct
}

func (c *structValueContainer) IsReadOnly() bool {
	return c.st == nil
}

func (c *structValueContainer) Self() any {
	return c.st
}

func (c *structValueContainer) get(key string) (*structpb.Value, error) {
	fs := c.st.GetFields()
	if fs == nil {
		return nil, protopatch.ErrNotFound{Kind: "key", Value: key}
	}
	v, ok := fs[key]
	if !ok {
		return nil, protopatch.ErrNotFound{Kind: "key", Value: key}
	}
	return v, nil
}

func (c *structValueContainer) Get(key string) (any, error) {
	return c.get(key)
}

func (c *structValueContainer) GetCopy(key string) (any, error) {
	v, err := c.get(key)
	if err != nil {
		return nil, err
	}
	return proto.Clone(v), nil
}

func (c *structValueContainer) GetNew(key string) (any, error) {
	return &structpb.Value{}, nil
}

func (c *structValueContainer) Mutable(key string) (any, error) {
	if c.st == nil {
		return nil, protopatch.ErrMutationOfReadOnlyValue
	}
	v, err := c.get(key)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (c *structValueContainer) Access(key string) (protopatch.Container, error) {
	v, err := c.get(key)
	if err != nil {
		return nil, err
	}
	return &valueContainer{v: v}, nil
}

func (c *structValueContainer) AccessMutable(key string) (protopatch.Container, error) {
	if c.st == nil {
		return nil, protopatch.ErrMutationOfReadOnlyValue
	}
	v, err := c.get(key)
	if err != nil {
		return nil, err
	}
	return &valueContainer{v: v}, nil
}

func (c *structValueContainer) Set(key string, to any) error {
	if c.st == nil {
		return protopatch.ErrMutationOfReadOnlyValue
	}
	if to == nil {
		delete(c.st.Fields, key)
		return nil
	}
	pr := c.st.ProtoReflect()
	if err := protoops.SetMessageFieldMapItem(pr, pr.Descriptor().Fields().ByName("fields"), protoreflect.ValueOf(key).MapKey(), to); err != nil {
		return protopatch.NewErrInPath(key, protopatch.ErrOperationFailed{Op: "set", Cause: err})
	}
	return nil
}

func (c *structValueContainer) Append(_ any) error {
	if c.st == nil {
		return protopatch.ErrMutationOfReadOnlyValue
	}
	return protopatch.ErrAppendToNonList
}

func (c *structValueContainer) Insert(_ string, _ any) error {
	if c.st == nil {
		return protopatch.ErrMutationOfReadOnlyValue
	}
	return protopatch.ErrAppendToNonList
}

type listValueContainer struct {
	li *structpb.ListValue
}

func (c *listValueContainer) IsReadOnly() bool { return c.li == nil }

func (c *listValueContainer) Self() any { return c.li }

func (c *listValueContainer) get(index string) (*structpb.Value, error) {
	vs := c.li.GetValues()
	idx := protoops.ParsedIndexInList(len(vs), index)
	if idx < 0 {
		return nil, protopatch.ErrNotFound{Kind: "index", Value: index}
	}
	return vs[idx], nil
}

func (c *listValueContainer) Get(index string) (any, error) {
	return c.get(index)
}

func (c *listValueContainer) GetCopy(index string) (any, error) {
	v, err := c.get(index)
	if err != nil {
		return nil, err
	}
	return proto.Clone(v), nil
}

func (c *listValueContainer) GetNew(index string) (any, error) {
	_, ok := protoops.ParseListIndex(index)
	if !ok {
		return nil, protopatch.ErrNotFound{Kind: "index", Value: index}
	}
	return &structpb.Value{}, nil
}

func (c *listValueContainer) Mutable(index string) (any, error) {
	if c.li == nil {
		return nil, protopatch.ErrMutationOfReadOnlyValue
	}
	v, err := c.get(index)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (c *listValueContainer) Access(index string) (protopatch.Container, error) {
	v, err := c.get(index)
	if err != nil {
		return nil, err
	}
	return &valueContainer{v: v}, nil
}

func (c *listValueContainer) AccessMutable(index string) (protopatch.Container, error) {
	if c.li == nil {
		return nil, protopatch.ErrMutationOfReadOnlyValue
	}
	v, err := c.get(index)
	if err != nil {
		return nil, err
	}
	return &valueContainer{v: v}, nil
}

func (c *listValueContainer) Set(index string, to any) error {
	if c.li == nil {
		return protopatch.ErrMutationOfReadOnlyValue
	}
	idx := protoops.ParsedIndexInList(len(c.li.Values), index)
	if idx < 0 {
		return protopatch.ErrNotFound{Kind: "index", Value: index}
	}
	if to == nil {
		c.li.Values = removeSliceIndex(c.li.Values, idx)
		return nil
	}
	pr := c.li.ProtoReflect()
	if err := protoops.SetMessageFieldListItem(pr, pr.Descriptor().Fields().ByName("values"), idx, to); err != nil {
		return protopatch.NewErrInPath(index, protopatch.ErrOperationFailed{Op: "set", Cause: err})
	}
	return nil
}

func (c *listValueContainer) Append(new any) error {
	if c.li == nil {
		return protopatch.ErrMutationOfReadOnlyValue
	}
	if new == nil {
		c.li.Values = append(c.li.Values, &structpb.Value{})
		return nil
	}
	pr := c.li.ProtoReflect()
	if err := protoops.AppendMessageFieldListItem(pr, pr.Descriptor().Fields().ByName("values"), new); err != nil {
		return protopatch.ErrOperationFailed{Op: "append", Cause: err}
	}
	return nil
}

func (c *listValueContainer) Insert(index string, new any) error {
	if c.li == nil {
		return protopatch.ErrMutationOfReadOnlyValue
	}
	idx := protoops.ParsedIndexInList(len(c.li.Values)+1, index)
	if idx < 0 {
		return protopatch.ErrNotFound{Kind: "index", Value: index}
	}
	if new == nil {
		c.li.Values = insertSliceIndex(c.li.Values, idx, &structpb.Value{})
		return nil
	}
	pr := c.li.ProtoReflect()
	if err := protoops.InsertMessageFieldListItem(pr, pr.Descriptor().Fields().ByName("values"), idx, new); err != nil {
		return protopatch.NewErrInPath(index, protopatch.ErrOperationFailed{Op: "insert", Cause: err})
	}
	return nil
}

type valueContainer struct {
	v *structpb.Value
}

func (c *valueContainer) IsReadOnly() bool { return c.v == nil }

func (c *valueContainer) Self() any { return c.v }

func (c *valueContainer) get(key string) (*structpb.Value, error) {
	switch k := c.v.GetKind().(type) {
	case *structpb.Value_StructValue:
		c := structValueContainer{st: k.StructValue}
		return c.get(key)
	case *structpb.Value_ListValue:
		c := listValueContainer{li: k.ListValue}
		return c.get(key)
	default:
		return nil, protopatch.ErrAccessToNonContainer
	}
}

func (c *valueContainer) Get(key string) (any, error) {
	return c.get(key)
}

func (c *valueContainer) GetCopy(key string) (any, error) {
	v, err := c.get(key)
	if err != nil {
		return nil, err
	}
	return proto.Clone(v), nil
}

func (c *valueContainer) GetNew(key string) (any, error) {
	switch k := c.v.GetKind().(type) {
	case *structpb.Value_StructValue:
		c := structValueContainer{st: k.StructValue}
		return c.GetNew(key)
	case *structpb.Value_ListValue:
		c := listValueContainer{li: k.ListValue}
		return c.GetNew(key)
	default:
		return nil, protopatch.ErrAccessToNonContainer
	}
}

func (c *valueContainer) Mutable(key string) (any, error) {
	if c.v == nil {
		return nil, protopatch.ErrMutationOfReadOnlyValue
	}
	v, err := c.get(key)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (c *valueContainer) Access(key string) (protopatch.Container, error) {
	switch k := c.v.GetKind().(type) {
	case *structpb.Value_StructValue:
		c := structValueContainer{st: k.StructValue}
		return c.Access(key)
	case *structpb.Value_ListValue:
		c := listValueContainer{li: k.ListValue}
		return c.Access(key)
	default:
		return nil, protopatch.ErrAccessToNonContainer
	}
}

func (c *valueContainer) AccessMutable(key string) (protopatch.Container, error) {
	if c.v == nil {
		return nil, protopatch.ErrMutationOfReadOnlyValue
	}
	switch k := c.v.GetKind().(type) {
	case *structpb.Value_StructValue:
		c := structValueContainer{st: k.StructValue}
		return c.Access(key) // use Access instead of AccessMutable in case k.StructValue is nil
	case *structpb.Value_ListValue:
		c := listValueContainer{li: k.ListValue}
		return c.Access(key) // use Access instead of AccessMutable in case k.ListValue is nil
	default:
		return nil, protopatch.ErrAccessToNonContainer
	}
}

func (c *valueContainer) Set(key string, to any) error {
	if c.v == nil {
		return protopatch.ErrMutationOfReadOnlyValue
	}
	switch k := c.v.GetKind().(type) {
	case *structpb.Value_StructValue:
		if k.StructValue == nil {
			k.StructValue = &structpb.Struct{}
		}
		c := structValueContainer{st: k.StructValue}
		return c.Set(key, to)
	case *structpb.Value_ListValue:
		if k.ListValue == nil {
			k.ListValue = &structpb.ListValue{}
		}
		c := listValueContainer{li: k.ListValue}
		return c.Set(key, to)
	default:
		return protopatch.ErrAccessToNonContainer
	}
}

func (c *valueContainer) Append(new any) error {
	if c.v == nil {
		return protopatch.ErrMutationOfReadOnlyValue
	}
	switch k := c.v.GetKind().(type) {
	case *structpb.Value_ListValue:
		if k.ListValue == nil {
			k.ListValue = &structpb.ListValue{}
		}
		c := listValueContainer{li: k.ListValue}
		return c.Append(new)
	default:
		return protopatch.ErrAccessToNonContainer
	}
}

func (c *valueContainer) Insert(key string, new any) error {
	if c.v == nil {
		return protopatch.ErrMutationOfReadOnlyValue
	}
	switch k := c.v.GetKind().(type) {
	case *structpb.Value_ListValue:
		if k.ListValue == nil {
			k.ListValue = &structpb.ListValue{}
		}
		c := listValueContainer{li: k.ListValue}
		return c.Insert(key, new)
	default:
		return protopatch.ErrAccessToNonContainer
	}
}
