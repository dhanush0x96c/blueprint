package ui

import (
	"errors"

	"github.com/dhanush0x96c/blueprint/internal/app"
)

// ExitCode returns an exit code for a given error.
func ExitCode(err error) int {
	switch {
	case errors.Is(err, app.ErrTemplateNotFound):
		return 2
	default:
		return 1
	}
}
