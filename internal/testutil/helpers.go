package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

func WriteTemplate(t testing.TB, dir string, content string) {
	t.Helper()

	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		t.Fatalf("failed to create template directory %s: %v", dir, err)
	}

	path := filepath.Join(dir, template.FileName)
	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("failed to write template file %s: %v", path, err)
	}
}

func NewRenderer(t testing.TB) (*template.Renderer, string) {
	t.Helper()

	return template.NewRenderer(), t.TempDir()
}

func Context(vars map[string]any) *template.Context {
	return &template.Context{
		Variables: vars,
	}
}
