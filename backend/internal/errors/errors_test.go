package errors

import (
	"fmt"
	"testing"
)

func TestE(t *testing.T) {
	const op Op = "someOp.foo"
	kind := KindDatabaseError
	e1 := fmt.Errorf("Some downstack error")
	msg := "A message"

	err := E(op, kind, e1, msg).(*Error)

	if err.Kind != kind {
		t.Errorf("Error Kind not properly set.  Expected %v, found %v", kind, err.Kind)
	}

	if err.Op != op {
		t.Errorf("Error Op not properly set. Expected %v, found %v", op, err.Op)
	}

	if err.Err != e1 {
		t.Errorf("Error Err not properly set.  Expected %v, found %v", e1, err.Err)
	}

	if err.Msg != msg {
		t.Errorf("Error Msg not properly set. Expected:\n\"%v\"\nFound:\"%v\"", msg, err.Msg)
	}

	// panic case
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("E with bad arg did not panic")
		}
	}()

	// should panic!
	_ = E(t)
}

func TestKind(t *testing.T) {
	// Case where arg is not one of our errors
	err := fmt.Errorf("Some error")
	if Kind(err) != KindUnexpected {
		t.Errorf("Expected KindUnexpected, got code: %v", Kind(err))
	}

	err = E(KindDatabaseError, fmt.Errorf("Database Error"))
	if Kind(err) != KindDatabaseError {
		t.Errorf("Expected KindDatabaseError, got code: %v", Kind(err))
	}

	// Kind is a couple layers deep
	err = E(E(E(KindNotFound, fmt.Errorf("Not found"))))
	if Kind(err) != KindNotFound {
		t.Errorf("Expected KindNotFound, got code: %v", Kind(err))
	}
}

func TestOps(t *testing.T) {
	err := E(Op("Foo"), fmt.Errorf("Some error"))

	err = E(err, Op("Bar"))
	err = E(err, Op("Baz"))

	ops := Ops(err)

	if ops[0] != Op("Baz") {
		t.Fail()
	}

	if ops[1] != Op("Bar") {
		t.Fail()
	}

	if ops[2] != Op("Foo") {
		t.Fail()
	}

	// Not our error case
	if len(Ops(fmt.Errorf("error"))) != 0 {
		t.Errorf("Ops length should be zero for outside errors")
	}
}

func TestError_Error(t *testing.T) {
	// Test with msg
	err := E("Some context", fmt.Errorf("underlying error"))
	if err.Error() != "Some context: underlying error" {
		t.Errorf("Error() string incorrect, got:\n%v", err.Error())
	}

	// Test without msg

	err = fmt.Errorf("underlying error")
	if err.Error() != E(err).Error() {
		t.Errorf("Error() message incorrect when no msg provided")
	}
}