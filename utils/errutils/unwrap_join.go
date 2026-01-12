package errutils

import "errors"

// UnwrapJoinError unwraps an error returned by errors.Join.
// If the error does not implement Unwrap() []error, it returns a slice
// containing the original error.
//
// Example:
//
//	err1 := errors.New("first error")
//	err2 := errors.New("second error")
//	joinedErr := errors.Join(err1, err2)
//
//	unwrapped, ok := UnwrapJoinError(joinedErr)
//	if ok {
//	    // unwrapped is []error{err1, err2}
//	}
func UnwrapJoinError(err error) ([]error, bool) {
	if err == nil {
		return nil, false
	}

	var we interface{ Unwrap() []error }
	if errors.As(err, &we) {
		return we.Unwrap(), true
	}

	return []error{err}, true
}
