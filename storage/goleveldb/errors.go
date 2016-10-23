package goleveldb

import "fmt"

type initError struct {
	cause error
}

type notFoundError struct {
	cause error
}

func (e initError) Error() string {
	return fmt.Sprintf("init: %s", e.cause.Error())
}

func (e initError) Cause() error {
	return e.cause
}

func (e initError) IsInitializer() bool {
	return true
}

func (e notFoundError) Error() string {
	return fmt.Sprintf("not found: %s", e.cause.Error())
}

func (e notFoundError) Cause() error {
	return e.cause
}

func (e notFoundError) IsNotFound() bool {
	return true
}
