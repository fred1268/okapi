package testing

import "errors"

var (
	// ErrStatusCodeMismatched is returned if HTTP Status code
	// differs from expected during a test.
	ErrStatusCodeMismatched error = errors.New("status code mismatched")
	// ErrResponseMismatched is returned if the server returned a
	// content that differs from expected during a test.
	ErrResponseMismatched error = errors.New("response mismatched")
	// ErrInvalidServerConfiguration is returned if the server
	// configuration is not valid.
	ErrInvalidServerConfiguration error = errors.New("invalid server configuration")
)
