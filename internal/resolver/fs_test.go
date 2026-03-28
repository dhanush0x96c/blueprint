package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/stretchr/testify/require"
)

func writeTemplate(t *testing.T, dir string, content string) {
	t.Helper()

	err := os.MkdirAll(dir, 0o755)
	require.NoError(t, err)

	path := filepath.Join(dir, template.FileName)
	err = os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)
}

const validProjectTemplate = `
name: go-cli
type: project
version: "1.0.0"
description: "Go CLI project"
variables:
  - name: app_name
    prompt: "App name?"
    type: string
    role: project_name
`

const validFeatureTemplate = `
name: testing
type: feature
version: "1.0.0"
description: "Testing support"
`

const invalidTemplate = `
name:
type: project
`

func TestFSResolver_Exists(t *testing.T) {
	base := t.TempDir()
	resolver := NewFSResolver(os.DirFS(base))

	templateName := "exists"
	dir := filepath.Join(base, templateName)
	writeTemplate(t, dir, validProjectTemplate)

	require.True(t, resolver.Exists(templateName))
	require.False(t, resolver.Exists("missing"))
}

func TestFSResolver_Discover(t *testing.T) {
	base := t.TempDir()
	resolver := NewFSResolver(os.DirFS(base))

	writeTemplate(t, filepath.Join(base, "projects", "go-cli"), validProjectTemplate)
	writeTemplate(t, filepath.Join(base, "features", "testing"), validFeatureTemplate)
	writeTemplate(t, filepath.Join(base, "broken"), invalidTemplate)

	templates, err := resolver.Discover()
	require.NoError(t, err)
	require.Len(t, templates, 2)
	require.Contains(t, templates, "projects/go-cli")
	require.Contains(t, templates, "features/testing")
	require.Equal(t, "go-cli", templates["projects/go-cli"].Name)
	require.Equal(t, template.TypeFeature, templates["features/testing"].Type)
}
