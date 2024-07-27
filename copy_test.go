package protopatch_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/daishe/protopatch"
	"github.com/daishe/protopatch/internal/patchtest"
	protopatchv1 "github.com/daishe/protopatch/internal/testtypes/protopatch/v1"
)

func TestCopy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		base            proto.Message
		targetPath      string
		replacementPath string
		opts            []protopatch.Option
		want            proto.Message
		wantErr         error
	}{
		{
			name:            "scalar/from/message/scalar",
			base:            &protopatchv1.TestMessage{String_: "aaa", Message: &protopatchv1.TestMessage{String_: "bbb"}},
			targetPath:      "string",
			replacementPath: "message.string",
			want:            &protopatchv1.TestMessage{String_: "bbb", Message: &protopatchv1.TestMessage{String_: "bbb"}},
		},

		{
			name:            "scalar-list/from/message/scalar-list",
			base:            &protopatchv1.TestMessage{List: &protopatchv1.TestList{String_: []string{"aaa"}}, Message: &protopatchv1.TestMessage{List: &protopatchv1.TestList{String_: []string{"bbb"}}}},
			targetPath:      "list.string",
			replacementPath: "message.list.string",
			want:            &protopatchv1.TestMessage{List: &protopatchv1.TestList{String_: []string{"bbb"}}, Message: &protopatchv1.TestMessage{List: &protopatchv1.TestList{String_: []string{"bbb"}}}},
		},
		{
			name:            "scalar-list/item/from/scalar-list/item",
			base:            &protopatchv1.TestList{String_: []string{"aaa", "bbb"}},
			targetPath:      "string.0",
			replacementPath: "string.1",
			want:            &protopatchv1.TestList{String_: []string{"bbb", "bbb"}},
		},

		{
			name:            "scalar-map/from/message/scalar-map",
			base:            &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}}, Message: &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToString: map[string]string{"key": "bbb"}}}},
			targetPath:      "map.stringToString",
			replacementPath: "message.map.stringToString",
			want:            &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToString: map[string]string{"key": "bbb"}}, Message: &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToString: map[string]string{"key": "bbb"}}}},
		},
		{
			name:            "scalar-map/item/from/scalar-map/item",
			base:            &protopatchv1.TestMap{StringToString: map[string]string{"key0": "aaa", "key1": "bbb"}},
			targetPath:      "stringToString.key0",
			replacementPath: "stringToString.key1",
			want:            &protopatchv1.TestMap{StringToString: map[string]string{"key0": "bbb", "key1": "bbb"}},
		},

		{
			name:            "base/from/message",
			base:            &protopatchv1.TestMessage{String_: "aaa", Message: &protopatchv1.TestMessage{String_: "bbb"}},
			targetPath:      "",
			replacementPath: "message",
			want:            &protopatchv1.TestMessage{String_: "bbb"},
		},

		{
			name:            "message/from/message/message",
			base:            &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{String_: "aaa", Message: &protopatchv1.TestMessage{String_: "bbb"}}},
			targetPath:      "message",
			replacementPath: "message.message",
			want:            &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{String_: "bbb"}},
		},

		{
			name:            "message-list/from/message/message-list",
			base:            &protopatchv1.TestMessage{List: &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}}, Message: &protopatchv1.TestMessage{List: &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}}}},
			targetPath:      "list.message",
			replacementPath: "message.list.message",
			want:            &protopatchv1.TestMessage{List: &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}}, Message: &protopatchv1.TestMessage{List: &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}}}},
		},
		{
			name:            "message-list/item/from/message-list/item",
			base:            &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}, {String_: "bbb"}}},
			targetPath:      "message.0",
			replacementPath: "message.1",
			want:            &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}, {String_: "bbb"}}},
		},

		{
			name:            "message-map/from/message/message-map",
			base:            &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}}, Message: &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "bbb"}}}}},
			targetPath:      "map.stringToMessage",
			replacementPath: "message.map.stringToMessage",
			want:            &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "bbb"}}}, Message: &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "bbb"}}}}},
		},
		{
			name:            "message-map/item/from/message-map/item",
			base:            &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key0": {String_: "aaa"}, "key1": {String_: "bbb"}}},
			targetPath:      "stringToMessage.key0",
			replacementPath: "stringToMessage.key1",
			want:            &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key0": {String_: "bbb"}, "key1": {String_: "bbb"}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			base := proto.Clone(test.base)
			err := protopatch.Copy(base, test.targetPath, test.replacementPath, test.opts...)

			if test.wantErr != nil {
				require.Equal(t, test.wantErr, err)
				return
			}
			require.NoError(t, err)
			patchtest.RequireEqual(t, test.want, base, "copy value mismatch")
		})
	}
}
