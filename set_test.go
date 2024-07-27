package protopatch_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/daishe/protopatch"
	"github.com/daishe/protopatch/internal/patchtest"
	protopatchv1 "github.com/daishe/protopatch/internal/testtypes/protopatch/v1"
)

func TestSet(t *testing.T) {
	t.Parallel()

	// convertErrOpt is a special conversion option that in case of a source value implements the error interface converter returns this error - useful for testing conversion return value propagation
	convertErrOpt := protopatch.WithConversion(protopatch.ConverterFunc(func(to, from any) (any, error) {
		if err, ok := from.(error); ok {
			return nil, err
		}
		return nil, protopatch.ErrNoConversionDefined
	}))

	mustAccessSelf := func(base proto.Message, path string) any {
		c, err := protopatch.Access(protopatch.MessageContainer(base), protopatch.Path(path))
		require.NoError(t, err)
		return c.Self()
	}

	tests := []struct {
		name    string
		base    proto.Message
		path    string
		value   any
		opts    []protopatch.Option
		want    proto.Message
		wantErr error
	}{
		{
			name:  "base/noop",
			base:  &protopatchv1.TestMessage{},
			path:  "",
			value: &protopatchv1.TestMessage{},
			want:  &protopatchv1.TestMessage{},
		},
		{
			name:  "base/set",
			base:  &protopatchv1.TestMessage{},
			path:  "",
			value: &protopatchv1.TestMessage{String_: "zzz"},
			want:  &protopatchv1.TestMessage{String_: "zzz"},
		},
		{
			name:    "base/set-wrong-type",
			base:    &protopatchv1.TestMessage{},
			path:    "",
			value:   123,
			wantErr: protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType},
		},
		{
			name:    "base/set-wrong-descriptor",
			base:    &protopatchv1.TestMessage{},
			path:    "",
			value:   &protopatchv1.TestOneof{},
			wantErr: protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType},
		},
		{
			name:    "nil-base/set",
			base:    (*protopatchv1.TestMessage)(nil),
			path:    "",
			value:   &protopatchv1.TestMessage{String_: "zzz"},
			wantErr: protopatch.ErrMutationOfReadOnlyValue,
		},

		{
			name:  "scalar/noop",
			base:  &protopatchv1.TestMessage{String_: "aaa"},
			path:  "string",
			value: "aaa",
			want:  &protopatchv1.TestMessage{String_: "aaa"},
		},
		{
			name:  "scalar/set",
			base:  &protopatchv1.TestMessage{String_: "aaa"},
			path:  "string",
			value: "bbb",
			want:  &protopatchv1.TestMessage{String_: "bbb"},
		},
		{
			name:    "scalar/set-wrong-type",
			base:    &protopatchv1.TestMessage{String_: "aaa"},
			path:    "string",
			value:   123,
			wantErr: protopatch.NewErrInPath("string", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:  "oneof/unset-scalar/set",
			base:  &protopatchv1.TestOneof{},
			path:  "string",
			value: "bbb",
			want:  &protopatchv1.TestOneof{Types: &protopatchv1.TestOneof_String_{String_: "bbb"}},
		},
		{
			name:  "oneof/set-scalar/set",
			base:  &protopatchv1.TestOneof{Types: &protopatchv1.TestOneof_String_{String_: "aaa"}},
			path:  "string",
			value: "bbb",
			want:  &protopatchv1.TestOneof{Types: &protopatchv1.TestOneof_String_{String_: "bbb"}},
		},
		{
			name:  "scalar-list/noop",
			base:  &protopatchv1.TestList{String_: []string{"aaa"}},
			path:  "string",
			value: []string{"aaa"},
			want:  &protopatchv1.TestList{String_: []string{"aaa"}},
		},
		{
			name:  "scalar-list/set",
			base:  &protopatchv1.TestList{String_: []string{"aaa"}},
			path:  "string",
			value: []string{"bbb"},
			want:  &protopatchv1.TestList{String_: []string{"bbb"}},
		},
		{
			name:  "scalar-list/set/from-protopatch-list",
			base:  &protopatchv1.TestList{String_: []string{"aaa"}},
			path:  "string",
			value: mustAccessSelf(&protopatchv1.TestList{String_: []string{"bbb"}}, "string"),
			want:  &protopatchv1.TestList{String_: []string{"bbb"}},
		},
		{
			name:  "scalar-list/set/from-read-only-protopatch-list",
			base:  &protopatchv1.TestList{String_: []string{"aaa"}},
			path:  "string",
			value: mustAccessSelf((*protopatchv1.TestList)(nil), "string"),
			want:  &protopatchv1.TestList{String_: []string{}},
		},
		{
			name:    "scalar-list/set-wrong-type",
			base:    &protopatchv1.TestList{String_: []string{"aaa"}},
			path:    "string",
			value:   []int32{123},
			wantErr: protopatch.NewErrInPath("string", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:  "scalar-list/item/noop",
			base:  &protopatchv1.TestList{String_: []string{"aaa"}},
			path:  "string.0",
			value: "aaa",
			want:  &protopatchv1.TestList{String_: []string{"aaa"}},
		},
		{
			name:  "scalar-list/item/set",
			base:  &protopatchv1.TestList{String_: []string{"aaa"}},
			path:  "string.0",
			value: "bbb",
			want:  &protopatchv1.TestList{String_: []string{"bbb"}},
		},
		{
			name:    "scalar-list/item/set-wrong-type",
			base:    &protopatchv1.TestList{String_: []string{"aaa"}},
			path:    "string.0",
			value:   123,
			wantErr: protopatch.NewErrInPath("string.0", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:    "scalar-list/unknown-item",
			base:    &protopatchv1.TestList{String_: []string{"aaa"}},
			path:    "string.1",
			value:   "bbb",
			wantErr: protopatch.NewErrInPath("string", protopatch.ErrNotFound{Kind: "index", Value: "1"}),
		},
		{
			name:  "scalar-map/noop",
			base:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}},
			path:  "stringToString",
			value: map[string]string{"key": "aaa"},
			want:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}},
		},
		{
			name:  "scalar-map/set",
			base:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}},
			path:  "stringToString",
			value: map[string]string{"key": "bbb"},
			want:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "bbb"}},
		},
		{
			name:  "scalar-map/set/from-protopatch-map",
			base:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}},
			path:  "stringToString",
			value: mustAccessSelf(&protopatchv1.TestMap{StringToString: map[string]string{"key": "bbb"}}, "stringToString"),
			want:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "bbb"}},
		},
		{
			name:  "scalar-map/set/from-read-only-protopatch-map",
			base:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}},
			path:  "stringToString",
			value: mustAccessSelf((*protopatchv1.TestMap)(nil), "stringToString"),
			want:  &protopatchv1.TestMap{StringToString: map[string]string{}},
		},
		{
			name:    "scalar-map/set-wrong-type",
			base:    &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}},
			path:    "stringToString",
			value:   map[string]int32{"key": 123},
			wantErr: protopatch.NewErrInPath("stringToString", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:  "scalar-map/item/noop",
			base:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}},
			path:  "stringToString.key",
			value: "aaa",
			want:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}},
		},
		{
			name:  "scalar-map/item/set",
			base:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}},
			path:  "stringToString.key",
			value: "bbb",
			want:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "bbb"}},
		},
		{
			name:    "scalar-map/item/set-wrong-type",
			base:    &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}},
			path:    "stringToString.key",
			value:   123,
			wantErr: protopatch.NewErrInPath("stringToString.key", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:  "scalar-map/unknown-item/set",
			base:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}},
			path:  "stringToString.unknown",
			value: "bbb",
			want:  &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa", "unknown": "bbb"}},
		},

		{
			name:  "message/noop",
			base:  &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{String_: "aaa"}},
			path:  "message",
			value: &protopatchv1.TestMessage{String_: "aaa"},
			want:  &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{String_: "aaa"}},
		},
		{
			name:  "message/set",
			base:  &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{String_: "aaa"}},
			path:  "message",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{String_: "bbb"}},
		},
		{
			name:    "message/set-wrong-type",
			base:    &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{String_: "aaa"}},
			path:    "message",
			value:   123,
			wantErr: protopatch.NewErrInPath("message", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:    "message/set-wrong-descriptor",
			base:    &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{String_: "aaa"}},
			path:    "message",
			value:   &protopatchv1.TestList{},
			wantErr: protopatch.NewErrInPath("message", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:    "message/set-unknown-field",
			base:    &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{String_: "aaa"}},
			path:    "unknown",
			value:   &protopatchv1.TestMessage{String_: "bbb"},
			wantErr: protopatch.ErrNotFound{Kind: "field", Value: "unknown"},
		},
		{
			name:  "oneof/unset-message/set",
			base:  &protopatchv1.TestOneof{},
			path:  "message",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestOneof{Types: &protopatchv1.TestOneof_Message{Message: &protopatchv1.TestMessage{String_: "bbb"}}},
		},
		{
			name:  "oneof/set-message/set",
			base:  &protopatchv1.TestOneof{Types: &protopatchv1.TestOneof_Message{Message: &protopatchv1.TestMessage{String_: "aaa"}}},
			path:  "message",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestOneof{Types: &protopatchv1.TestOneof_Message{Message: &protopatchv1.TestMessage{String_: "bbb"}}},
		},
		{
			name:  "message-list/noop",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:  "message",
			value: []*protopatchv1.TestMessage{{String_: "aaa"}},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
		},
		{
			name:  "message-list/set",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:  "message",
			value: []*protopatchv1.TestMessage{{String_: "bbb"}},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}},
		},
		{
			name:  "message-list/set/from-protopatch-list",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:  "message",
			value: mustAccessSelf(&protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}}, "message"),
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}},
		},
		{
			name:  "message-list/set/from-read-only-protopatch-list",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:  "message",
			value: mustAccessSelf((*protopatchv1.TestList)(nil), "message"),
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{}},
		},
		{
			name:    "message-list/set-wrong-type",
			base:    &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:    "message",
			value:   []int32{123},
			wantErr: protopatch.NewErrInPath("message", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:    "message-list/set-wrong-descriptor",
			base:    &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:    "message",
			value:   []*protopatchv1.TestOneof{{}},
			wantErr: protopatch.NewErrInPath("message", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:  "message-list/item/noop",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:  "message.0",
			value: &protopatchv1.TestMessage{String_: "aaa"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
		},
		{
			name:  "message-list/item/set",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:  "message.0",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}},
		},
		{
			name:    "message-list/item/set-wrong-type",
			base:    &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:    "message.0",
			value:   123,
			wantErr: protopatch.NewErrInPath("message.0", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:    "message-list/item/set-wrong-descriptor",
			base:    &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:    "message.0",
			value:   &protopatchv1.TestOneof{},
			wantErr: protopatch.NewErrInPath("message.0", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:    "message-list/unknown-item/set",
			base:    &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:    "message.1",
			value:   &protopatchv1.TestMessage{String_: "bbb"},
			wantErr: protopatch.NewErrInPath("message", protopatch.ErrNotFound{Kind: "index", Value: "1"}),
		},
		{
			name:  "message-map/noop",
			base:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
			path:  "stringToMessage",
			value: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}},
			want:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
		},
		{
			name:  "message-map/set",
			base:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
			path:  "stringToMessage",
			value: map[string]*protopatchv1.TestMessage{"key": {String_: "bbb"}},
			want:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "bbb"}}},
		},
		{
			name:  "message-map/set/from-protopatch-map",
			base:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
			path:  "stringToMessage",
			value: mustAccessSelf(&protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "bbb"}}}, "stringToMessage"),
			want:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "bbb"}}},
		},
		{
			name:  "message-map/set/from-read-only-protopatch-map",
			base:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
			path:  "stringToMessage",
			value: mustAccessSelf((*protopatchv1.TestMap)(nil), "stringToMessage"),
			want:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{}},
		},
		{
			name:    "message-map/set-wrong-type",
			base:    &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
			path:    "stringToMessage",
			value:   map[string]int32{"key": 123},
			wantErr: protopatch.NewErrInPath("stringToMessage", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:    "message-map/set-wrong-descriptor",
			base:    &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
			path:    "stringToMessage",
			value:   map[string]*protopatchv1.TestOneof{"key": {}},
			wantErr: protopatch.NewErrInPath("stringToMessage", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:  "message-map/item/noop",
			base:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
			path:  "stringToMessage.key",
			value: &protopatchv1.TestMessage{String_: "aaa"},
			want:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
		},
		{
			name:  "message-map/item/set",
			base:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
			path:  "stringToMessage.key",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "bbb"}}},
		},
		{
			name:    "message-map/item/set-wrong-type",
			base:    &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
			path:    "stringToMessage.key",
			value:   123,
			wantErr: protopatch.NewErrInPath("stringToMessage.key", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:    "message-map/item/set-wrong-descriptor",
			base:    &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
			path:    "stringToMessage.key",
			value:   &protopatchv1.TestOneof{},
			wantErr: protopatch.NewErrInPath("stringToMessage.key", protopatch.ErrOperationFailed{Op: "set", Cause: protopatch.ErrMismatchingType}),
		},
		{
			name:  "message-map/unknown-item/set",
			base:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}},
			path:  "stringToMessage.unknown",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}, "unknown": {String_: "bbb"}}},
		},

		{
			name:    "message/unknown-field/unknown",
			base:    &protopatchv1.TestMessage{},
			path:    "unknown.otherUnknown",
			value:   nil,
			wantErr: protopatch.ErrNotFound{Kind: "field", Value: "unknown"},
		},
		{
			name:    "base/convert-failure",
			base:    &protopatchv1.TestMessage{},
			path:    "",
			value:   errors.New("test error"),
			opts:    []protopatch.Option{convertErrOpt},
			wantErr: errors.New("test error"),
		},
		{
			name:    "message/scalar/convert-failure",
			base:    &protopatchv1.TestMessage{String_: "aaa"},
			path:    "string",
			value:   errors.New("test error"),
			opts:    []protopatch.Option{convertErrOpt},
			wantErr: protopatch.NewErrInPath("string", errors.New("test error")),
		},
		{
			name:    "message/message/scalar/convert-failure",
			base:    &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{String_: "aaa"}},
			path:    "message.string",
			value:   errors.New("test error"),
			opts:    []protopatch.Option{convertErrOpt},
			wantErr: protopatch.NewErrInPath("message.string", errors.New("test error")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			base := proto.Clone(test.base)
			err := protopatch.Set(base, test.path, test.value, test.opts...)

			if test.wantErr != nil {
				require.Equal(t, test.wantErr, err)
				return
			}
			require.NoError(t, err)
			patchtest.RequireEqual(t, test.want, base, "set value mismatch")
		})
	}
}

func TestContainerSetFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		base    proto.Message
		path    string
		key     string
		value   any
		wantErr error
	}{
		{
			name:    "read-only-message",
			base:    &protopatchv1.TestMessage{},
			path:    "message",
			key:     "string",
			value:   "aaa",
			wantErr: protopatch.ErrMutationOfReadOnlyValue,
		},
		{
			name:    "read-only-map",
			base:    &protopatchv1.TestMessage{},
			path:    "map.stringToString",
			key:     "key",
			value:   "aaa",
			wantErr: protopatch.ErrMutationOfReadOnlyValue,
		},
		{
			name:    "read-only-list",
			base:    &protopatchv1.TestMessage{},
			path:    "list.string",
			key:     "0",
			value:   "aaa",
			wantErr: protopatch.ErrMutationOfReadOnlyValue,
		},

		{
			name:    "message/unknown-field",
			base:    &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{}},
			path:    "message",
			key:     "unknown",
			value:   "aaa",
			wantErr: protopatch.ErrNotFound{Kind: "field", Value: "unknown"},
		},
		{
			name:    "list/unknown-index",
			base:    &protopatchv1.TestMessage{List: &protopatchv1.TestList{String_: []string{}}},
			path:    "list.string",
			key:     "0",
			value:   "aaa",
			wantErr: protopatch.ErrNotFound{Kind: "index", Value: "0"},
		},
		{
			name:    "map/invalid-item",
			base:    &protopatchv1.TestMap{Int32ToString: map[int32]string{}},
			path:    "int32ToString",
			key:     "unknown",
			value:   "aaa",
			wantErr: protopatch.ErrNotFound{Kind: "key", Value: "unknown"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			base := proto.Clone(test.base)
			container, err := protopatch.Access(protopatch.MessageContainer(base), protopatch.Path(test.path))
			require.NoError(t, err)

			err = container.Set(test.key, test.value)
			require.Equal(t, test.wantErr, err)
		})
	}
}
