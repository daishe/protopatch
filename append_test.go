package protopatch_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/daishe/protopatch"
	"github.com/daishe/protopatch/internal/patchtest"
	protopatchv1 "github.com/daishe/protopatch/internal/testtypes/protopatch/v1"
)

func TestAppend(t *testing.T) {
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
			name:  "scalar-list-nil/append",
			base:  &protopatchv1.TestList{String_: []string(nil)},
			path:  "string",
			value: "bbb",
			want:  &protopatchv1.TestList{String_: []string{"bbb"}},
		},
		{
			name:  "scalar-list-empty/append",
			base:  &protopatchv1.TestList{String_: []string{}},
			path:  "string",
			value: "bbb",
			want:  &protopatchv1.TestList{String_: []string{"bbb"}},
		},
		{
			name:  "scalar-list/append",
			base:  &protopatchv1.TestList{String_: []string{"aaa"}},
			path:  "string",
			value: "bbb",
			want:  &protopatchv1.TestList{String_: []string{"aaa", "bbb"}},
		},
		{
			name:  "message-list-nil/append",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage(nil)},
			path:  "message",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}},
		},
		{
			name:  "message-list-empty/append",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{}},
			path:  "message",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "bbb"}}},
		},
		{
			name:  "message-list/append",
			base:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}}},
			path:  "message",
			value: &protopatchv1.TestMessage{String_: "bbb"},
			want:  &protopatchv1.TestList{Message: []*protopatchv1.TestMessage{{String_: "aaa"}, {String_: "bbb"}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			base := proto.Clone(test.base)
			err := protopatch.Append(base, test.path, test.value, test.opts...)

			if test.wantErr != nil {
				require.Equal(t, test.wantErr, err)
				return
			}
			require.NoError(t, err)
			patchtest.RequireEqual(t, test.want, base, "append value mismatch")
		})
	}
}
