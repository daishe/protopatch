package protopatch_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/daishe/protopatch"
	"github.com/daishe/protopatch/internal/patchtest"
	protopatchv1 "github.com/daishe/protopatch/internal/testtypes/protopatch/v1"
)

func TestInsert(t *testing.T) {
	t.Parallel()

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
			name:  "scalar-list-nil/insert-at-zero",
			base:  &protopatchv1.TestList{String_: []string(nil)},
			path:  "string.0",
			value: "bbb",
			want:  &protopatchv1.TestList{String_: []string{"bbb"}},
		},
		{
			name:  "scalar-list-nil/insert-at-negative-one",
			base:  &protopatchv1.TestList{String_: []string(nil)},
			path:  "string.-1",
			value: "bbb",
			want:  &protopatchv1.TestList{String_: []string{"bbb"}},
		},
		{
			name:  "scalar-list-empty/insert-at-zero",
			base:  &protopatchv1.TestList{String_: []string{}},
			path:  "string.0",
			value: "bbb",
			want:  &protopatchv1.TestList{String_: []string{"bbb"}},
		},
		{
			name:  "scalar-list-empty/insert-at-negative-one",
			base:  &protopatchv1.TestList{String_: []string{}},
			path:  "string.-1",
			value: "bbb",
			want:  &protopatchv1.TestList{String_: []string{"bbb"}},
		},
		{
			name:  "scalar-list/insert-at-zero",
			base:  &protopatchv1.TestList{String_: []string{"aaa"}},
			path:  "string.0",
			value: "bbb",
			want:  &protopatchv1.TestList{String_: []string{"bbb", "aaa"}},
		},
		{
			name:  "scalar-list/insert-at-one",
			base:  &protopatchv1.TestList{String_: []string{"aaa"}},
			path:  "string.1",
			value: "bbb",
			want:  &protopatchv1.TestList{String_: []string{"aaa", "bbb"}},
		},
		{
			name:  "scalar-list/insert-at-negative-one",
			base:  &protopatchv1.TestList{String_: []string{"aaa"}},
			path:  "string.-1",
			value: "bbb",
			want:  &protopatchv1.TestList{String_: []string{"aaa", "bbb"}},
		},
		{
			name:  "scalar-list/insert-at-negative-two",
			base:  &protopatchv1.TestList{String_: []string{"aaa"}},
			path:  "string.-2",
			value: "bbb",
			want:  &protopatchv1.TestList{String_: []string{"bbb", "aaa"}},
		},

		{
			name:  "message-list-nil/insert-at-zero",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage(nil)},
			path:  "message.0",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}},
		},
		{
			name:  "message-list-nil/insert-at-negative-one",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage(nil)},
			path:  "message.-1",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}},
		},
		{
			name:  "message-list-empty/insert-at-zero",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{}},
			path:  "message.0",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}},
		},
		{
			name:  "message-list-empty/insert-at-negative-one",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{}},
			path:  "message.-1",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}},
		},
		{
			name:  "message-list/insert-at-zero",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:  "message.0",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}, {String_: "aaa"}}},
		},
		{
			name:  "message-list/insert-at-one",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:  "message.1",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}, {String_: "bbb"}}},
		},
		{
			name:  "message-list/insert-at-negative-one",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:  "message.-1",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}, {String_: "bbb"}}},
		},
		{
			name:  "message-list/insert-at-negative-two",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:  "message.-2",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}, {String_: "aaa"}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			base := proto.Clone(test.base)
			err := protopatch.Insert(base, test.path, test.value, test.opts...)

			if test.wantErr != nil {
				require.Equal(t, test.wantErr, err)
				return
			}
			require.NoError(t, err)
			patchtest.RequireEqual(t, test.want, base, "insert value mismatch")
		})
	}
}
