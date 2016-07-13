package wttypes

import "errors"

var (
	ErrInvalidArgument = errors.New("invalid argument")

	ErrNotFound = errors.New("not found")

	ErrBadRoute = errors.New("bad route")
)


