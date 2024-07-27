package protopatch

import (
	"errors"
)

type Option interface {
	configure(*setup)
}

type optionFunc func(s *setup)

func (fn optionFunc) configure(s *setup) { fn(s) }

var ErrNoConversionDefined = errors.New("no conversion defined for the provided types")

// Converter represents an entity that can change type of one entity to another.
type Converter interface {
	// Convert performs type conversion to type of first provided item from the second provided value. Returned value must be of first item type or function should return an error. If conversion is not defined for provided types Convert function should return an unwrapped ErrNoConversionDefined error.
	Convert(to, from any) (any, error)
}

// ConverterFunc allows to implement Converter interface with a function.
type ConverterFunc func(to, from any) (any, error)

func (fn ConverterFunc) Convert(to, from any) (any, error) { return fn(to, from) }

func WithConversion(converters ...Converter) Option {
	return optionFunc(func(s *setup) {
		s.convert = converters // TODO: Copy slice to avoid referencing the passed value (?)
	})
}

var ErrNoContainerTransformationDefined = errors.New("no transformation defined for the provided container")

// ContainerTransformer represents an entity that can modify freshly accessed container.
type ContainerTransformer interface {
	// TransformContainer performs container transformation as needed. Returned value must be a valid container or function should return an error. If transformation is not defined for the given container function should return an unwrapped ErrNoContainerTransformationDefined error.
	TransformContainer(container Container) (Container, error)
}

// ContainerTransformerFunc allows to implement ContainerTransformer interface with a function.
type ContainerTransformerFunc func(container Container) (Container, error)

func (fn ContainerTransformerFunc) TransformContainer(container Container) (Container, error) {
	return fn(container)
}

func WithContainerTransformation(transformers ...ContainerTransformer) Option {
	return optionFunc(func(s *setup) {
		s.transform = transformers // TODO: Copy slice to avoid referencing the passed value (?)
	})
}

type setup struct {
	convert   []Converter
	transform []ContainerTransformer
}

func newSetup(opts ...Option) *setup {
	s := &setup{}
	for _, o := range opts {
		o.configure(s)
	}
	return s
}

func (s *setup) Convert(to, from any) (any, error) {
	for _, c := range s.convert {
		v, err := c.Convert(to, from)
		if err == ErrNoConversionDefined {
			continue
		}
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return IdentityConverter(to, from)
}

func convert(to, from any, setup *setup) (any, error) {
	conv, err := setup.Convert(to, from)
	if err == ErrNoConversionDefined {
		conv, err = from, nil
	}
	return conv, err
}

func (s *setup) TransformContainer(c Container) (Container, error) {
	for _, t := range s.transform {
		n, err := t.TransformContainer(c)
		if err == ErrNoContainerTransformationDefined {
			continue
		}
		if err != nil {
			return nil, err
		}
		return n, nil
	}
	return nil, ErrNoContainerTransformationDefined
}

func transformContainer(c Container, setup *setup) (Container, error) {
	n, err := setup.TransformContainer(c)
	if err == ErrNoContainerTransformationDefined {
		n, err = c, nil
	}
	return n, err
}
