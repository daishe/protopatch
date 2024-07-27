package protopatch_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daishe/protopatch"
)

func TestPath(t *testing.T) {
	t.Parallel()

	invalidSegmentInfo := segmentInfo{value: "", isFirst: true, isLast: true}

	tests := []struct {
		path     string
		segments []segmentInfo
	}{
		{
			path: "",
			segments: []segmentInfo{
				{value: "", isFirst: true, isLast: true},
			},
		},
		{
			path: "a",
			segments: []segmentInfo{
				{value: "a", isFirst: true, isLast: true, precedingPathWithCurrentSegment: "a", followingPathWithCurrentSegment: "a"},
			},
		},
		{
			path: "aa",
			segments: []segmentInfo{
				{value: "aa", isFirst: true, isLast: true, precedingPathWithCurrentSegment: "aa", followingPathWithCurrentSegment: "aa"},
			},
		},
		{
			path: "aaa",
			segments: []segmentInfo{
				{value: "aaa", isFirst: true, isLast: true, precedingPathWithCurrentSegment: "aaa", followingPathWithCurrentSegment: "aaa"},
			},
		},
		{
			path: "a.",
			segments: []segmentInfo{
				{value: "a", isFirst: true, precedingPathWithCurrentSegment: "a", followingPathWithCurrentSegment: "a."},
				{value: "", isLast: true, precedingPath: "a", precedingPathWithCurrentSegment: "a."},
			},
		},
		{
			path: "a.b",
			segments: []segmentInfo{
				{value: "a", isFirst: true, followingPath: "b", precedingPathWithCurrentSegment: "a", followingPathWithCurrentSegment: "a.b"},
				{value: "b", isLast: true, precedingPath: "a", precedingPathWithCurrentSegment: "a.b", followingPathWithCurrentSegment: "b"},
			},
		},
		{
			path: "a.b.",
			segments: []segmentInfo{
				{value: "a", isFirst: true, followingPath: "b.", precedingPathWithCurrentSegment: "a", followingPathWithCurrentSegment: "a.b."},
				{value: "b", precedingPath: "a", precedingPathWithCurrentSegment: "a.b", followingPathWithCurrentSegment: "b."},
				{value: "", isLast: true, precedingPath: "a.b", precedingPathWithCurrentSegment: "a.b."},
			},
		},
		{
			path: "a.b.c",
			segments: []segmentInfo{
				{value: "a", isFirst: true, followingPath: "b.c", precedingPathWithCurrentSegment: "a", followingPathWithCurrentSegment: "a.b.c"},
				{value: "b", precedingPath: "a", followingPath: "c", precedingPathWithCurrentSegment: "a.b", followingPathWithCurrentSegment: "b.c"},
				{value: "c", isLast: true, precedingPath: "a.b", precedingPathWithCurrentSegment: "a.b.c", followingPathWithCurrentSegment: "c"},
			},
		},
		{
			path: "a.b.c.",
			segments: []segmentInfo{
				{value: "a", isFirst: true, followingPath: "b.c.", precedingPathWithCurrentSegment: "a", followingPathWithCurrentSegment: "a.b.c."},
				{value: "b", precedingPath: "a", followingPath: "c.", precedingPathWithCurrentSegment: "a.b", followingPathWithCurrentSegment: "b.c."},
				{value: "c", precedingPath: "a.b", precedingPathWithCurrentSegment: "a.b.c", followingPathWithCurrentSegment: "c."},
				{value: "", isLast: true, precedingPath: "a.b.c", precedingPathWithCurrentSegment: "a.b.c."},
			},
		},
		{
			path: ".a",
			segments: []segmentInfo{
				{value: "", isFirst: true, followingPath: "a", followingPathWithCurrentSegment: ".a"},
				{value: "a", isLast: true, precedingPathWithCurrentSegment: ".a", followingPathWithCurrentSegment: "a"},
			},
		},
		{
			path: ".a.b",
			segments: []segmentInfo{
				{value: "", isFirst: true, followingPath: "a.b", followingPathWithCurrentSegment: ".a.b"},
				{value: "a", followingPath: "b", precedingPathWithCurrentSegment: ".a", followingPathWithCurrentSegment: "a.b"},
				{value: "b", isLast: true, precedingPath: ".a", precedingPathWithCurrentSegment: ".a.b", followingPathWithCurrentSegment: "b"},
			},
		},
		{
			path: ".a.",
			segments: []segmentInfo{
				{value: "", isFirst: true, followingPath: "a.", followingPathWithCurrentSegment: ".a."},
				{value: "a", precedingPathWithCurrentSegment: ".a", followingPathWithCurrentSegment: "a."},
				{value: "", isLast: true, precedingPath: ".a", precedingPathWithCurrentSegment: ".a."},
			},
		},
		{
			path: ".",
			segments: []segmentInfo{
				{value: "", isFirst: true, followingPathWithCurrentSegment: "."},
				{value: "", isLast: true, precedingPathWithCurrentSegment: "."},
			},
		},
		{
			path: "..",
			segments: []segmentInfo{
				{value: "", isFirst: true, followingPath: ".", followingPathWithCurrentSegment: ".."},
				{value: "", precedingPathWithCurrentSegment: ".", followingPathWithCurrentSegment: "."},
				{value: "", isLast: true, precedingPath: ".", precedingPathWithCurrentSegment: ".."},
			},
		},
	}

	for _, test := range tests {
		t.Run("path-"+test.path+"-", func(t *testing.T) {
			t.Parallel()

			path := protopatch.Path(test.path)

			test.segments[0].requireEqual(t, path.First(), "path first segment")
			test.segments[len(test.segments)-1].requireEqual(t, path.Last(), "path last segment")

			require.Equal(t, len(test.segments), path.SegmentsCount())
			segments := path.Segments()
			require.Equal(t, len(test.segments), len(segments))

			i := 0
			for ps := range path.Iter {
				test.segments[i].requireEqual(t, ps, fmt.Sprintf("segment #%d from iterator", i))
				test.segments[i].requireEqual(t, segments[i], fmt.Sprintf("segment #%d from slice", i))
				if i > 0 {
					test.segments[i-1].requireEqual(t, ps.Previous(), fmt.Sprintf("previous of segment #%d from iterator", i))
					test.segments[i-1].requireEqual(t, segments[i].Previous(), fmt.Sprintf("previous of segment #%d from slice", i))
				} else {
					invalidSegmentInfo.requireEqual(t, ps.Previous(), fmt.Sprintf("previous of segment #%d from iterator", i))
					invalidSegmentInfo.requireEqual(t, segments[i].Previous(), fmt.Sprintf("previous of segment #%d from slice", i))
				}
				if i < len(test.segments)-1 {
					test.segments[i+1].requireEqual(t, ps.Next(), fmt.Sprintf("next to segment #%d from iterator", i))
					test.segments[i+1].requireEqual(t, segments[i].Next(), fmt.Sprintf("next to segment #%d from slice", i))
				} else {
					invalidSegmentInfo.requireEqual(t, ps.Next(), fmt.Sprintf("next to segment #%d from iterator", i))
					invalidSegmentInfo.requireEqual(t, segments[i].Next(), fmt.Sprintf("next to segment #%d from slice", i))
				}
				i++
			}
			require.Equal(t, len(test.segments), i)
		})
	}
}

type segmentInfo struct {
	precedingPath                   string
	precedingPathWithCurrentSegment string
	value                           string
	followingPath                   string
	followingPathWithCurrentSegment string
	isFirst                         bool
	isLast                          bool
}

func (si segmentInfo) requireEqual(t *testing.T, ps protopatch.PathSegment, msg string) {
	require.Equal(t, si.value, ps.Value(), "invalid Value method result, "+msg)
	require.Equal(t, si.isLast, ps.IsLast(), "invalid IsLast method result, "+msg)
	require.Equal(t, si.isFirst, ps.IsFirst(), "invalid IsFirst method result, "+msg)
	require.Equal(t, si.precedingPath, string(ps.PrecedingPath()), "invalid PrecedingPath method result, "+msg)
	require.Equal(t, si.precedingPathWithCurrentSegment, string(ps.PrecedingPathWithCurrentSegment()), "invalid PrecedingPathWithCurrentSegment method result, "+msg)
	require.Equal(t, si.followingPath, string(ps.FollowingPath()), "invalid FollowingPath method result, "+msg)
	require.Equal(t, si.followingPathWithCurrentSegment, string(ps.FollowingPathWithCurrentSegment()), "invalid FollowingPathWithCurrentSegment method result, "+msg)
}

func TestPathJoin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		path  string
		given []string
		want  string
	}{
		{
			name:  "no-op-empty",
			path:  "",
			given: []string{},
			want:  "",
		},
		{
			name:  "no-op-value",
			path:  "a",
			given: []string{},
			want:  "a",
		},
		{
			name:  "empty-with-empty",
			path:  "",
			given: []string{""},
			want:  ".",
		},
		{
			name:  "value-with-empty",
			path:  "a",
			given: []string{""},
			want:  "a.",
		},
		{
			name:  "empty-with-value",
			path:  "",
			given: []string{"a"},
			want:  ".a",
		},
		{
			name:  "multiple-values",
			path:  "a",
			given: []string{"b", "c"},
			want:  "a.b.c",
		},
		{
			name:  "dots-with-values",
			path:  "..",
			given: []string{"a", "b"},
			want:  "...a.b",
		},
		{
			name:  "values-with-dots",
			path:  "a.b",
			given: []string{".", ".."},
			want:  "a.b.....",
		},
	}

	for _, test := range tests {
		t.Run("Join/"+test.name, func(t *testing.T) {
			t.Parallel()
			other := make([]protopatch.Path, 0, len(test.given))
			for _, g := range test.given {
				other = append(other, protopatch.Path(g))
			}
			got := protopatch.Path(test.path).Join(other...)
			require.Equal(t, protopatch.Path(test.want), got)
		})
		t.Run("JoinSegment/"+test.name, func(t *testing.T) {
			t.Parallel()
			other := make([]protopatch.PathSegment, 0, len(test.given))
			for _, g := range test.given {
				other = append(other, protopatch.Path(g).Segments()...)
			}
			got := protopatch.Path(test.path).JoinSegment(other...)
			require.Equal(t, protopatch.Path(test.want), got)
		})
		t.Run("JoinSegmentValue/"+test.name, func(t *testing.T) {
			t.Parallel()
			got := protopatch.Path(test.path).JoinSegmentValue(test.given...)
			require.Equal(t, protopatch.Path(test.want), got)
		})
	}
}
