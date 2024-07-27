package protopatch_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/daishe/protopatch"
	"github.com/daishe/protopatch/internal/patchtest"
	protopatchv1 "github.com/daishe/protopatch/internal/testtypes/protopatch/v1"
)

func TestSwap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		base       proto.Message
		firstPath  string
		secondPath string
		opts       []protopatch.Option
		want       proto.Message
		wantErr    error
	}{
		{
			name:       "scalar/with/message/scalar",
			base:       &protopatchv1.TestMessage{String_: "aaa", Message: &protopatchv1.TestMessage{String_: "bbb"}},
			firstPath:  "string",
			secondPath: "message.string",
			want:       &protopatchv1.TestMessage{String_: "bbb", Message: &protopatchv1.TestMessage{String_: "aaa"}},
		},

		{
			name:       "scalar-list/with/message/scalar-list",
			base:       &protopatchv1.TestMessage{List: &protopatchv1.TestList{String_: []string{"aaa"}}, Message: &protopatchv1.TestMessage{List: &protopatchv1.TestList{String_: []string{"bbb"}}}},
			firstPath:  "list.string",
			secondPath: "message.list.string",
			want:       &protopatchv1.TestMessage{List: &protopatchv1.TestList{String_: []string{"bbb"}}, Message: &protopatchv1.TestMessage{List: &protopatchv1.TestList{String_: []string{"aaa"}}}},
		},
		{
			name:       "scalar-list/item/with/scalar-list/item",
			base:       &protopatchv1.TestList{String_: []string{"aaa", "bbb"}},
			firstPath:  "string.0",
			secondPath: "string.1",
			want:       &protopatchv1.TestList{String_: []string{"bbb", "aaa"}},
		},

		{
			name:       "scalar-map/with/message/scalar-map",
			base:       &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}}, Message: &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToString: map[string]string{"key": "bbb"}}}},
			firstPath:  "map.stringToString",
			secondPath: "message.map.stringToString",
			want:       &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToString: map[string]string{"key": "bbb"}}, Message: &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToString: map[string]string{"key": "aaa"}}}},
		},
		{
			name:       "scalar-map/item/with/scalar-map/item",
			base:       &protopatchv1.TestMap{StringToString: map[string]string{"key0": "aaa", "key1": "bbb"}},
			firstPath:  "stringToString.key0",
			secondPath: "stringToString.key1",
			want:       &protopatchv1.TestMap{StringToString: map[string]string{"key0": "bbb", "key1": "aaa"}},
		},

		{
			name:       "base/with/message",
			base:       &protopatchv1.TestMessage{String_: "aaa", Message: &protopatchv1.TestMessage{String_: "bbb"}},
			firstPath:  "",
			secondPath: "message",
			want:       &protopatchv1.TestMessage{String_: "bbb"},
		},

		{
			name:       "message/with/message/message",
			base:       &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{String_: "aaa", Message: &protopatchv1.TestMessage{String_: "bbb"}}},
			firstPath:  "message",
			secondPath: "message.message",
			want:       &protopatchv1.TestMessage{Message: &protopatchv1.TestMessage{String_: "bbb"}},
		},

		{
			name:       "message-list/with/message/message-list",
			base:       &protopatchv1.TestMessage{List: &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}}, Message: &protopatchv1.TestMessage{List: &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}}}},
			firstPath:  "list.message",
			secondPath: "message.list.message",
			want:       &protopatchv1.TestMessage{List: &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}}, Message: &protopatchv1.TestMessage{List: &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}}}},
		},
		{
			name:       "message-list/item/with/message-list/item",
			base:       &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}, {String_: "bbb"}}},
			firstPath:  "message.0",
			secondPath: "message.1",
			want:       &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}, {String_: "aaa"}}},
		},

		{
			name:       "message-map/with/message/message-map",
			base:       &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}}, Message: &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "bbb"}}}}},
			firstPath:  "map.stringToMessage",
			secondPath: "message.map.stringToMessage",
			want:       &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "bbb"}}}, Message: &protopatchv1.TestMessage{Map: &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key": {String_: "aaa"}}}}},
		},
		{
			name:       "message-map/item/with/message-map/item",
			base:       &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key0": {String_: "aaa"}, "key1": {String_: "bbb"}}},
			firstPath:  "stringToMessage.key0",
			secondPath: "stringToMessage.key1",
			want:       &protopatchv1.TestMap{StringToMessage: map[string]*protopatchv1.TestMessage{"key0": {String_: "bbb"}, "key1": {String_: "aaa"}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			base := proto.Clone(test.base)
			err := protopatch.Swap(base, test.firstPath, test.secondPath, test.opts...)

			if test.wantErr != nil {
				require.Equal(t, test.wantErr, err)
				return
			}
			require.NoError(t, err)
			patchtest.RequireEqual(t, test.want, base, "swap value mismatch")
		})
	}
}
