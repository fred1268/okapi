package testing

import "errors"

var (
	ErrStatusCodeMismatched error = errors.New("status code mismatched")
	ErrResultMismatched     error = errors.New("result mismatched")
	ErrInvalidJWTSession    error = errors.New("invalid JWT session")
)
