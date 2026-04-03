package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// Writer handles writing files and directories to disk
type Writer struct {
	defaultPerm os.FileMode
	dirPerm     os.FileMode
}

// WriteResult contains the files written and skipped during a write operation.
type WriteResult struct {
	Written []string
	Skipped []string
}

// NewWriter creates a new file writer with default permissions
func NewWriter() *Writer {
	return &Writer{
		defaultPerm: 0644, // rw-r--r--
		dirPerm:     0755, // rwxr-xr-x
	}
}

// WriteFile writes content to a file, creating parent directories if needed
func (w *Writer) WriteFile(path string, content []byte) error {
	return w.WriteFileWithPerm(path, content, w.defaultPerm)
}

// WriteFiles writes multiple rendered files into the given output directory.
func (w *Writer) WriteFiles(outputDir string, files []template.RenderedFile, overwrite bool) (*WriteResult, error) {
	result := &WriteResult{
		Written: make([]string, 0, len(files)),
		Skipped: make([]string, 0),
	}

	for _, file := range files {
		fullPath := filepath.Join(outputDir, file.Path)

		if _, err := os.Stat(fullPath); err == nil && !overwrite {
			result.Skipped = append(result.Skipped, file.Path)
			continue
		}

		if err := w.WriteFile(fullPath, file.Content); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}

		result.Written = append(result.Written, file.Path)
	}

	return result, nil
}

// WriteFileWithPerm writes content to a file with specific permissions
func (w *Writer) WriteFileWithPerm(path string, content []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := w.EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, content, perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// EnsureDir creates a directory and all parent directories if they don't exist
func (w *Writer) EnsureDir(path string) error {
	if path == "" || path == "." {
		return nil
	}

	if err := os.MkdirAll(path, w.dirPerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}
