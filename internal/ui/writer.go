package ui

import (
	"fmt"
	"io"
)

func write(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func writeln(w io.Writer, s string) {
	_, _ = fmt.Fprintln(w, s)
}
