package ui

import (
	"errors"

	"github.com/dhanush0x96c/blueprint/internal/app"
)

// ExitCode returns an exit code for a given error.
func ExitCode(err error) int {
	var templateNotFoundErr *app.TemplateNotFoundError

	switch {
	case errors.As(err, &templateNotFoundErr):
		return 2
	default:
		return 1
	}
}
