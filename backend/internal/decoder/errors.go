package decoder

import "errors"

var (
	ErrValidation   = errors.New("invalid decoder configuration")
	ErrNotFound     = errors.New("decoder configuration not found")
	ErrConflict     = errors.New("decoder configuration comparison changed")
	ErrCVValidation = errors.New("invalid CV input")
	ErrCVConflict   = errors.New("vehicle CV changed or target identity already exists")
)
