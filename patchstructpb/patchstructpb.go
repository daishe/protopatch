package patchstructpb

type Option interface {
	configure(*setup)
}

type optionFunc func(*setup)

func (fn optionFunc) configure(s *setup) { fn(s) }

// IgnoreUnknownKeys returns option that ignores unknown keys when converting from structpb.Struct to proto.Message or when parsing a map key fails.
func IgnoreUnknownKeys() Option {
	return optionFunc(func(s *setup) {
		s.ignoreUnknownStructKeysForMessages = true
		s.ignoreUnknownStructKeysForMaps = true
		s.clearUnknownSourceStructKeys = false
		s.clearInvalidSourceValues = false
	})
}

// ClearUnknownKeys returns option that clears from the source structure each inner unknown key when converting from structpb.Struct to proto.Message or when parsing a map key fails.
func ClearUnknownKeys() Option {
	return optionFunc(func(s *setup) {
		s.ignoreUnknownStructKeysForMessages = false
		s.ignoreUnknownStructKeysForMaps = false
		s.clearUnknownSourceStructKeys = true
		s.clearInvalidSourceValues = true
	})
}

// IgnoreInvalidValues returns option that ignores each inner source value which type do not matches.
func IgnoreInvalidValues() Option {
	return optionFunc(func(s *setup) {
		s.ignoreInvalidValues = true
		s.clearInvalidSourceValues = false
	})
}

// ClearInvalidValues returns option that clears from the source structure each inner value which type do not matches.
func ClearInvalidValues() Option {
	return optionFunc(func(s *setup) {
		s.ignoreInvalidValues = false
		s.clearInvalidSourceValues = true
	})
}

type setup struct {
	ignoreUnknownStructKeysForMessages bool
	ignoreUnknownStructKeysForMaps     bool
	clearUnknownSourceStructKeys       bool
	ignoreInvalidValues                bool
	clearInvalidSourceValues           bool
	// convertFromInterface               bool
}

func newSetup(opts ...Option) *setup {
	s := &setup{}
	for _, o := range opts {
		o.configure(s)
	}
	return s
}
