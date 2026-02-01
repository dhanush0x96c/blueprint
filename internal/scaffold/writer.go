package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

// Writer handles writing files and directories to disk
type Writer struct {
	defaultPerm os.FileMode
	dirPerm     os.FileMode
}

// NewWriter creates a new file writer with default permissions
func NewWriter() *Writer {
	return &Writer{
		defaultPerm: 0644, // rw-r--r--
		dirPerm:     0755, // rwxr-xr-x
	}
}

// NewWriterWithPerms creates a new file writer with custom permissions
func NewWriterWithPerms(filePerm, dirPerm os.FileMode) *Writer {
	return &Writer{
		defaultPerm: filePerm,
		dirPerm:     dirPerm,
	}
}

// WriteFile writes content to a file, creating parent directories if needed
func (w *Writer) WriteFile(path string, content string) error {
	return w.WriteFileWithPerm(path, content, w.defaultPerm)
}

// WriteFileWithPerm writes content to a file with specific permissions
func (w *Writer) WriteFileWithPerm(path string, content string, perm os.FileMode) error {
	// Create parent directories
	dir := filepath.Dir(path)
	if err := w.EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// WriteFiles writes multiple files from a map of path -> content
func (w *Writer) WriteFiles(files map[string]string) error {
	for path, content := range files {
		if err := w.WriteFile(path, content); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}
	return nil
}

// WriteFilesWithBase writes multiple files with a base directory prefix
func (w *Writer) WriteFilesWithBase(baseDir string, files map[string]string) error {
	for path, content := range files {
		fullPath := filepath.Join(baseDir, path)
		if err := w.WriteFile(fullPath, content); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
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

// FileExists checks if a file exists at the given path
func (w *Writer) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists at the given path
func (w *Writer) DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsEmpty checks if a directory is empty
func (w *Writer) IsEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("failed to read directory: %w", err)
	}
	return len(entries) == 0, nil
}

// SafeWrite writes a file only if it doesn't exist
// Returns true if the file was written, false if it was skipped
func (w *Writer) SafeWrite(path string, content string) (bool, error) {
	if w.FileExists(path) {
		return false, nil
	}

	if err := w.WriteFile(path, content); err != nil {
		return false, err
	}

	return true, nil
}

// SafeWriteFiles writes multiple files, skipping existing ones
// Returns a list of files that were written and a list that were skipped
func (w *Writer) SafeWriteFiles(files map[string]string) (written []string, skipped []string, err error) {
	written = make([]string, 0)
	skipped = make([]string, 0)

	for path, content := range files {
		wasWritten, err := w.SafeWrite(path, content)
		if err != nil {
			return written, skipped, fmt.Errorf("failed to write file %s: %w", path, err)
		}

		if wasWritten {
			written = append(written, path)
		} else {
			skipped = append(skipped, path)
		}
	}

	return written, skipped, nil
}

// SafeWriteFilesWithBase is like SafeWriteFiles but with a base directory prefix
func (w *Writer) SafeWriteFilesWithBase(baseDir string, files map[string]string) (written []string, skipped []string, err error) {
	written = make([]string, 0)
	skipped = make([]string, 0)

	for path, content := range files {
		fullPath := filepath.Join(baseDir, path)
		wasWritten, err := w.SafeWrite(fullPath, content)
		if err != nil {
			return written, skipped, fmt.Errorf("failed to write file %s: %w", path, err)
		}

		if wasWritten {
			written = append(written, path)
		} else {
			skipped = append(skipped, path)
		}
	}

	return written, skipped, nil
}

// RemoveFile removes a file if it exists
func (w *Writer) RemoveFile(path string) error {
	if !w.FileExists(path) {
		return nil
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	return nil
}

// RemoveDir removes a directory and all its contents
func (w *Writer) RemoveDir(path string) error {
	if !w.DirExists(path) {
		return nil
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}

	return nil
}

// CopyFile copies a file from src to dst
func (w *Writer) CopyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	return w.WriteFile(dst, string(content))
}

// GetAbsolutePath returns the absolute path for a given path
func (w *Writer) GetAbsolutePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	return absPath, nil
}

// SetDefaultPerm sets the default file permission
func (w *Writer) SetDefaultPerm(perm os.FileMode) {
	w.defaultPerm = perm
}

// SetDirPerm sets the default directory permission
func (w *Writer) SetDirPerm(perm os.FileMode) {
	w.dirPerm = perm
}

// GetDefaultPerm returns the default file permission
func (w *Writer) GetDefaultPerm() os.FileMode {
	return w.defaultPerm
}

// GetDirPerm returns the default directory permission
func (w *Writer) GetDirPerm() os.FileMode {
	return w.dirPerm
}
