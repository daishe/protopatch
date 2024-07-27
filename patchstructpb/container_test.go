package patchstructpb_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/daishe/protopatch"
	"github.com/daishe/protopatch/internal/patchtest"
	protopatchv1 "github.com/daishe/protopatch/internal/testtypes/protopatch/v1"
	"github.com/daishe/protopatch/patchstructpb"
)

func TestValueContainerTransform(t *testing.T) {
	tests := []struct {
		name  string
		given proto.Message
		path  string
		want  any
	}{
		{
			name:  "access/string-value",
			given: &protopatchv1.TestWellKnown{Value: structpb.NewStringValue("aaa")},
			path:  "value",
			want:  structpb.NewStringValue("aaa"),
		},
		{
			name: "access/struct-value",
			given: &protopatchv1.TestWellKnown{
				Value: structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
					"key": structpb.NewStringValue("aaa"),
				}}),
			},
			path: "value",
			want: structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
				"key": structpb.NewStringValue("aaa"),
			}}),
		},
		{
			name: "struct-value/access/string-value",
			given: &protopatchv1.TestWellKnown{
				Value: structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
					"key": structpb.NewStringValue("aaa"),
				}}),
			},
			path: "value.key",
			want: structpb.NewStringValue("aaa"),
		},
		{
			name: "struct-value/struct-value/access/string-value",
			given: &protopatchv1.TestWellKnown{
				Value: structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
					"key0": structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
						"key1": structpb.NewStringValue("aaa"),
					}}),
				}}),
			},
			path: "value.key0.key1",
			want: structpb.NewStringValue("aaa"),
		},
		{
			name: "list-value/access/string-value",
			given: &protopatchv1.TestWellKnown{
				Value: structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{
					structpb.NewStringValue("aaa"),
				}}),
			},
			path: "value.0",
			want: structpb.NewStringValue("aaa"),
		},
		{
			name: "list-value/list-value/access/string-value",
			given: &protopatchv1.TestWellKnown{
				Value: structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{
					structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{
						structpb.NewStringValue("aaa"),
					}}),
				}}),
			},
			path: "value.0.0",
			want: structpb.NewStringValue("aaa"),
		},
	}
	for _, test := range tests {
		t.Run("read-only/"+test.name, func(t *testing.T) {
			t.Parallel()
			base := proto.Clone(test.given)
			c := protopatch.MessageContainer(base)

			// validate base has not been changed after access, as read-only operation should never modify the provided value
			defer patchtest.RequireEqual(t, test.given, base, "base modified after read-only access")

			accessed, err := protopatch.Access(c, protopatch.Path(test.path), protopatch.WithContainerTransformation(patchstructpb.ValueContainerTransformer()))
			require.NoError(t, err)

			patchtest.RequireEqual(t, test.want, accessed.Self(), "wrong self value after read-only access")
		})
	}
}
