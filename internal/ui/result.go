package ui

import (
	"os"

	"github.com/dhanush0x96c/blueprint/internal/scaffold"
)

// RenderResult prints a summary of the scaffolding result to stdout.
func RenderResult(result *scaffold.Result) {
	w := os.Stdout

	if len(result.FilesWritten) > 0 {
		writeln(w, "\nFiles written:")
		for _, f := range result.FilesWritten {
			write(w, "  ✓ %s\n", f)
		}
	}

	if len(result.FilesSkipped) > 0 {
		writeln(w, "\nFiles skipped (already exist):")
		for _, f := range result.FilesSkipped {
			write(w, "  - %s\n", f)
		}
	}

	if len(result.Dependencies) > 0 {
		writeln(w, "\nDependencies declared:")
		for _, dep := range result.Dependencies {
			write(w, "  • %s\n", dep)
		}
	}

	if len(result.PostInitCmds) > 0 {
		writeln(w, "\nPost-init commands:")
		for _, cmd := range result.PostInitCmds {
			write(w, "  $ %s\n", cmd.Command)
		}
	}

	if len(result.FilesWritten) == 0 && len(result.FilesSkipped) == 0 {
		writeln(w, "No files were written.")
	}
}
