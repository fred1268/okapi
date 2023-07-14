package testing

import "errors"

var (
	// ErrStatusCodeMismatched is returned if HTTP Status code
	// differs from expected during a test.
	ErrStatusCodeMismatched error = errors.New("status code mismatched")
	// ErrResultMismatched is returned if server returned content
	// differs from expected during a test.
	ErrResultMismatched error = errors.New("result mismatched")
	// ErrInvalidServerConfiguration is returned if the server
	// configuration is not valid.
	ErrInvalidServerConfiguration error = errors.New("invalid server configuration")
)