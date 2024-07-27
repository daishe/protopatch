package protopatch

import "strings"

const PathSegmentSeparator = "."

type PathSegment struct {
	path        Path
	startOffset int
	endOffset   int
}

func (ps PathSegment) PrecedingPath() Path {
	return ps.path[:max(0, ps.startOffset-1)]
}

func (ps PathSegment) PrecedingPathWithCurrentSegment() Path {
	return ps.path[:ps.endOffset]
}

func (ps PathSegment) Previous() PathSegment {
	if ps.startOffset == 0 {
		return PathSegment{}
	}
	end := ps.startOffset - 1
	start := strings.LastIndex(string(ps.path[0:end]), PathSegmentSeparator) + 1
	return PathSegment{path: ps.path, startOffset: start, endOffset: end}
}

func (ps PathSegment) Value() string {
	return string(ps.path[ps.startOffset:ps.endOffset])
}

func (ps PathSegment) Next() PathSegment {
	if ps.endOffset == len(ps.path) {
		return PathSegment{}
	}
	start := ps.endOffset + 1
	end := start + strings.Index(string(ps.path[start:]), PathSegmentSeparator)
	if end < start {
		end = len(ps.path)
	}
	return PathSegment{path: ps.path, startOffset: start, endOffset: end}
}

func (ps PathSegment) FollowingPath() Path {
	return ps.path[min(ps.endOffset+1, len(ps.path)):]
}

func (ps PathSegment) FollowingPathWithCurrentSegment() Path {
	return ps.path[ps.startOffset:]
}

func (ps PathSegment) IsFirst() bool {
	return ps.startOffset == 0
}

func (ps PathSegment) IsLast() bool {
	return ps.endOffset == len(ps.path)
}

type Path string

func (p Path) First() PathSegment {
	end := strings.Index(string(p), PathSegmentSeparator)
	if end < 0 {
		end = len(p)
	}
	return PathSegment{path: p, startOffset: 0, endOffset: end}
}

func (p Path) Last() PathSegment {
	return PathSegment{path: p, startOffset: strings.LastIndex(string(p), PathSegmentSeparator) + 1, endOffset: len(p)}
}

func (p Path) SegmentsCount() int {
	return strings.Count(string(p), PathSegmentSeparator) + 1
}

func (p Path) Segments() []PathSegment {
	s := make([]PathSegment, p.SegmentsCount())
	i := 0
	for ps := range p.Iter {
		s[i] = ps
		i++
	}
	return s
}

func (p Path) Iter(yield func(PathSegment) bool) {
	ps := p.First()
	for !ps.IsLast() {
		if !yield(ps) {
			return
		}
		ps = ps.Next()
	}
	yield(ps)
}

// Join create a new Path by concatenating all provided paths.
func (p Path) Join(other ...Path) Path {
	for _, o := range other {
		p += PathSegmentSeparator + o
	}
	return p
}

// JoinSegment create a new Path by concatenating all provided segments at the ent of the give Path.
func (p Path) JoinSegment(segments ...PathSegment) Path {
	for _, s := range segments {
		p += Path(PathSegmentSeparator + s.Value())
	}
	return p
}

// JoinSegmentValue create a new Path by concatenating all provided segment values at the ent of the give Path.
func (p Path) JoinSegmentValue(segments ...string) Path {
	for _, s := range segments {
		p += Path(PathSegmentSeparator + s)
	}
	return p
}
