package memory

import "fmt"

type initError struct {
	cause error
}

type notFoundError struct {
	cause error
}

// IsInitializer checks if the error value is an initializer error.
func IsInitializer(err error) bool {
	type init interface {
		IsInitializer() bool
	}

	if i, ok := err.(init); ok {
		return ok && i.IsInitializer()
	} else {
		return false
	}
}

// IsNotFound checks if the error value is a not-found error.
func IsNotFound(err error) bool {
	type notFound interface {
		IsNotFound() bool
	}

	if nf, ok := err.(notFound); ok {
		return ok && nf.IsNotFound()
	} else {
		return false
	}
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
