package ui

import (
	"os"

	"github.com/dhanush0x96c/blueprint/internal/app"
)

func renderTemplateNotFound(err *app.TemplateNotFoundError) {
	w := os.Stderr

	write(w, "✗ Template not found: %s\n", err.Name)
	writeln(w, "")
	writeln(w, "Hint:")
	writeln(w, "  Run `blueprint list` to see available templates.")
}

func renderInvalidTemplateType(err *app.InvalidTemplateTypeError) {
	w := os.Stderr

	write(w, "✗ Invalid template type: %s\n", err.Type)
	writeln(w, "")
	writeln(w, "Hint:")
	writeln(w, "  Valid types are: projects, features, components")
}
