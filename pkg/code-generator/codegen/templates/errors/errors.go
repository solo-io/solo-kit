package errors

import "errors"

var (
	UnrecognizedSourceType = errors.New("unrecognized source type; this should never happen")
)
