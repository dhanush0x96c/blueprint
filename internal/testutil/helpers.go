// Package testutil provides common test fakes, builders, and helpers across package tests.
package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/template/loader"
	"github.com/dhanush0x96c/blueprint/internal/template/renderer"
	"gopkg.in/yaml.v3"
)

func WriteTemplate(t testing.TB, dir string, content string) {
	t.Helper()

	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		t.Fatalf("failed to create template directory %s: %v", dir, err)
	}

	path := filepath.Join(dir, loader.FileName)
	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("failed to write template file %s: %v", path, err)
	}
}

func WriteTemplateStruct(t testing.TB, dir string, tmpl *template.Template) {
	t.Helper()

	data, err := yaml.Marshal(tmpl)
	if err != nil {
		t.Fatalf("failed to marshal template: %v", err)
	}

	WriteTemplate(t, dir, string(data))
}

func NewRenderer(t testing.TB) (*renderer.Renderer, string) {
	t.Helper()

	return renderer.NewRenderer(), t.TempDir()
}

func Context(vars map[string]any) *template.Context {
	return &template.Context{
		Variables: vars,
	}
}
