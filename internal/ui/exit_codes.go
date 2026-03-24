package ui

import (
	"errors"

	"github.com/dhanush0x96c/blueprint/internal/app"
)

// Exit codes as documented in docs/cli.md
const (
	ExitSuccess          = 0
	ExitGeneralError     = 1
	ExitInvalidArguments = 2
	ExitTemplateNotFound = 3
	ExitValidationFailed = 4
	ExitFilesystemError  = 5
	ExitInterrupted      = 130
)

// ExitCode returns an exit code for a given error.
func ExitCode(err error) int {
	var templateNotFoundErr *app.TemplateNotFoundError
	var invalidTemplateTypeErr *app.InvalidTemplateTypeError

	switch {
	case errors.As(err, &templateNotFoundErr):
		return ExitTemplateNotFound
	case errors.As(err, &invalidTemplateTypeErr):
		return ExitInvalidArguments
	default:
		return ExitGeneralError
	}
}
