package butil

import "errors"

func Coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}

type noCriticalError struct {
	err error
}

func (e *noCriticalError) Error() string {
	if errors.Is(e.err, nil) {
		return ""
	}

	return e.Error()
}

func (e *noCriticalError) Unwrap() error {
	return errors.Unwrap(e.err)
}

func NotCriticalError(err error) error {
	if errors.Is(err, nil) {
		return nil
	}

	return &noCriticalError{err: err}
}

func IsNotCriticalError(err error) bool {
	var target *noCriticalError

	return errors.As(err, &target)
}
