package ui

import (
	"errors"
	"os"

	"github.com/dhanush0x96c/blueprint/internal/app"
)

// RenderError dispatches the given error to the appropriate renderer based on its type.
func RenderError(err error) {
	var templateNotFoundErr *app.TemplateNotFoundError
	var invalidTemplateTypeErr *app.InvalidTemplateTypeError

	switch {
	case errors.As(err, &templateNotFoundErr):
		renderTemplateNotFound(templateNotFoundErr)
	case errors.As(err, &invalidTemplateTypeErr):
		renderInvalidTemplateType(invalidTemplateTypeErr)
	default:
		renderDefault(err)
	}
}

func renderDefault(err error) {
	write(os.Stderr, "error: %v\n", err)
}
