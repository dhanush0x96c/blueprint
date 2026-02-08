package ui

import (
	"errors"
	"os"

	"github.com/dhanush0x96c/blueprint/internal/app"
)

func RenderError(err error) {
	switch {
	case errors.Is(err, app.ErrTemplateNotFound):
		renderTemplateNotFound(err)
	default:
		renderDefault(err)
	}
}

func renderDefault(err error) {
	write(os.Stderr, "error: %v\n", err)
}
