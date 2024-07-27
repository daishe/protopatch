package protopatch

import (
	"errors"
	"fmt"

	"github.com/daishe/protopatch/internal/protoops"
)

var (
	ErrAccessToNonContainer = errors.New("connot descend into specified key; attempted to access a sub-value in an entity that is not a message, list or map")
	ErrAppendToNonList      = errors.New("connot append to non list field")
	ErrInsertToNonList      = errors.New("connot insert to non list field")
	// ErrDeleteNonKey            = errors.New("connot delete non key value")
	ErrMutationOfReadOnlyValue = errors.New("connot mutate read-only value")
	ErrMismatchingType         = protoops.ErrMismatchingType
)

type ErrNotFound struct {
	Kind  string // "field", "index" or "key"
	Value string
}

func (e ErrNotFound) Error() string {
	if e.Kind == "" {
		return fmt.Sprintf("%q not found", e.Value)
	}
	return fmt.Sprintf("%s %q not found", e.Kind, e.Value)
}

type ErrOperationFailed struct {
	Op    string
	Cause error
}

func newSetFailure(cause error) ErrOperationFailed {
	return ErrOperationFailed{Op: "set", Cause: cause}
}

func newAppendFailure(cause error) ErrOperationFailed {
	return ErrOperationFailed{Op: "append", Cause: cause}
}

func newInsertFailure(cause error) ErrOperationFailed {
	return ErrOperationFailed{Op: "insert", Cause: cause}
}

func (e ErrOperationFailed) Error() string {
	if e.Op == "" {
		return fmt.Sprintf("operation failed: %s", e.Cause.Error())
	}
	return fmt.Sprintf("cannot %s: %s", e.Op, e.Cause.Error())
}

func (e ErrOperationFailed) Unwrap() error {
	return e.Cause
}

// type ErrSwapMissmatchingType struct {
// 	FirstPath  string
// 	SecondPath string
// }

// func (e ErrSwapMissmatchingType) Error() string {
// 	return fmt.Sprintf("cannot swap %q with missmatching type %q", e.FirstPath, e.SecondPath)
// }

type ErrInPath struct {
	Path  string
	Cause error
}

func NewErrInPath(path string, err error) error {
	if err == nil {
		return nil
	}
	if pe, ok := err.(ErrInPath); ok { // updat path only on direct sub path error, do not use errors.As
		pe.Path = string(Path(path).Join(Path(pe.Path)))
		return pe
	}
	return ErrInPath{Path: path, Cause: err}
}

func (e ErrInPath) Error() string {
	if e.Path == "" {
		return e.Cause.Error()
	}
	return fmt.Sprintf("proto path %q: %s", e.Path, e.Cause.Error())
}

func (e ErrInPath) Unwrap() error {
	return e.Cause
}
