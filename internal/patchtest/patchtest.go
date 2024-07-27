package patchtest

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/daishe/protopatch"
)

func AssertEqual(t *testing.T, want, got any, msgAndArgs ...any) bool {
	t.Helper()
	if diff := cmp.Diff(want, got, transformList(), transformMap(), protocmp.Transform()); diff != "" {
		msg := formatMsgAndArgs(msgAndArgs...)
		if msg == "" {
			msg = "values mismatch"
		}
		t.Errorf("%s (-want +got):\n%s", msg, diff)
		return false
	}
	return true
}

func RequireEqual(t *testing.T, want, got any, msgAndArgs ...any) {
	t.Helper()
	if !AssertEqual(t, want, got, msgAndArgs...) {
		t.FailNow()
	}
}

func formatMsgAndArgs(msgAndArgs ...any) string {
	switch len(msgAndArgs) {
	case 0:
		return ""
	case 1:
		return msgAndArgs[0].(string)
	}
	return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
}

func addrType(t reflect.Type) reflect.Type {
	if k := t.Kind(); k == reflect.Interface || k == reflect.Ptr {
		return t
	}
	return reflect.PointerTo(t)
}

func pathImplements(p cmp.Path, t reflect.Type) bool {
	last := p.Last()
	if last.Type().Implements(t) {
		return true
	}
	if last.Type().Kind() == reflect.Interface {
		vx, vy := last.Values()
		if !vx.IsValid() || vx.IsNil() || !vy.IsValid() || vy.IsNil() {
			return false
		}
		return addrType(vx.Elem().Type()).Implements(t) || addrType(vy.Elem().Type()).Implements(t)
	}
	return false
}

var protopatchListType = reflect.TypeOf((*protopatch.List)(nil)).Elem()

func transformList() cmp.Option {
	return cmp.FilterPath(
		func(p cmp.Path) bool {
			return pathImplements(p, protopatchListType)
		},
		cmp.Transformer("ListToGoSlice", func(v any) any {
			if li, ok := v.(protopatch.List); ok {
				return li.AsGoSlice()
			}
			return v
		}),
	)
}

var protopatchMapType = reflect.TypeOf((*protopatch.Map)(nil)).Elem()

func transformMap() cmp.Option {
	return cmp.FilterPath(
		func(p cmp.Path) bool {
			return pathImplements(p, protopatchMapType)
		},
		cmp.Transformer("ListToGoSlice", func(v any) any {
			if ma, ok := v.(protopatch.Map); ok {
				return ma.AsGoMap()
			}
			return v
		}),
	)
}
