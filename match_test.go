package protopatch_test

// import (
// 	"reflect"
// 	"testing"

// 	protopatchv1 "github.com/daishe/protopatch/internal/testtypes/protopatch/v1"
// 	"google.golang.org/protobuf/proto"
// )

// func TestMatch(t *testing.T) {
// 	var a = []*protopatchv1.TestMessage{}
// 	var b []proto.Message

// 	t.Log("testing a")
// 	foo(t, a)
// 	t.Log("testing b")
// 	foo(t, b)
// 	t.Fail()
// }

// var messageType = reflect.TypeOf((*proto.Message)(nil)).Elem()

// func foo(t *testing.T, x any) {
// 	v := reflect.Zero(reflect.TypeOf(x).Elem())
// 	t.Logf("type=%v", v.Type().Name())

// 	if v.Type() == messageType {
// 		t.Logf("is message type")
// 	} else {
// 		t.Logf("is not message type")
// 	}

// 	if !v.IsValid() {
// 		t.Logf("invalid reflect value")
// 		return
// 	}
// 	// if v.IsNil() {
// 	// 	t.Logf("nil reflect value")
// 	// 	return
// 	// }

// 	m, ok := v.Interface().(proto.Message)
// 	if !ok {
// 		t.Logf("not a message")
// 		return
// 	}
// 	if m == nil {
// 		t.Logf("message is nil")
// 		return
// 	}
// 	pr := m.ProtoReflect()
// 	if pr == nil {
// 		t.Logf("proto reflect is nil")
// 		return
// 	}
// 	if pr.Descriptor() == nil {
// 		t.Logf("message descriptor is nil")
// 		return
// 	}
// 	t.Logf("no nils")
// }
