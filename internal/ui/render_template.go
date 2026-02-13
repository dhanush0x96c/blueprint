package ui

import (
	"os"
)

func renderTemplateNotFound(err error) {
	w := os.Stderr

	writeln(w, "✗ Template not found")
	writeln(w, "")
	writeln(w, "The requested template does not exist.")
	writeln(w, "")
	writeln(w, "Hint:")
	writeln(w, "  Run `blueprint list` to see available templates.")
}
