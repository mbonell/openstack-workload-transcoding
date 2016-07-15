package wttypes

import "errors"

var (
	// ErrInvalidArgument is used for "invalid argument"
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrNotFound is used for "not found"
	ErrNotFound = errors.New("not found")

	// ErrBadRoute is used for "bad route"
	ErrBadRoute = errors.New("bad route")

	ErrMismatchID = errors.New("Mistmach in URL ID and JSON ID")

	ErrNoJSON = errors.New("Response not in JSON format")

	ErrTranscodingNotFound = errors.New("Transcoding ID not found")


)
