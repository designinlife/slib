package errors_test

import (
	"testing"

	"github.com/designinlife/slib/errors"
)

type MyError struct {
	Msg string
}

func (e *MyError) Error() string {
	return e.Msg
}

func TestSlibErrors(t *testing.T) {
	// New / Newf
	err1 := errors.New("simple error")
	if err1.Error() != "simple error" {
		t.Errorf("New() got %q, want %q", err1.Error(), "simple error")
	}

	err2 := errors.Errorf("error %d", 42)
	if err2.Error() != "error 42" {
		t.Errorf("Newf() got %q, want %q", err2.Error(), "error 42")
	}

	// Wrap / Wrapf
	baseErr := errors.New("base")
	wrapped := errors.Wrap(baseErr, "context")
	if !errors.Is(wrapped, baseErr) {
		t.Errorf("Wrap() Is() failed: %v not wrapping %v", wrapped, baseErr)
	}
	if wrapped.Error() != "context: base" {
		t.Errorf("Wrap() got %q", wrapped.Error())
	}

	wrappedf := errors.Wrapf(baseErr, "context %d", 99)
	if !errors.Is(wrappedf, baseErr) {
		t.Errorf("Wrapf() Is() failed: %v not wrapping %v", wrappedf, baseErr)
	}
	if wrappedf.Error() != "context 99: base" {
		t.Errorf("Wrapf() got %q", wrappedf.Error())
	}

	// Is / As
	myErr := &MyError{"custom"}
	if !errors.As(myErr, &myErr) {
		t.Errorf("As() failed to match *MyError")
	}

	errChain := errors.Wrap(myErr, "outer")
	var target *MyError
	if !errors.As(errChain, &target) {
		t.Errorf("As() failed to unwrap *MyError from chain")
	}
	if target.Msg != "custom" {
		t.Errorf("As() got wrong unwrapped value: %v", target.Msg)
	}

	unwrapped := errors.Unwrap(wrapped)
	if unwrapped != baseErr {
		t.Errorf("Unwrap() got %v, want %v", unwrapped, baseErr)
	}
}
