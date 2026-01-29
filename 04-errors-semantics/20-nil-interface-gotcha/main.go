package main

import "fmt"

type MyError struct{ Op string }

func (e *MyError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("operation %s failed", e.Op)
}

func DoThing(fail bool) error {
	var e *MyError = nil
	if fail {
		e = &MyError{Op: "read"}
	}
	return e
}

func DoThingFix(fail bool) error {
	if fail {
		return &MyError{Op: "read"}
	}
	return nil
}

func WrapError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("wrapped: %w", err)
}
