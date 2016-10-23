package storage

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
