package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeTemplate(t *testing.T, dir string, content string) {
	t.Helper()

	err := os.MkdirAll(dir, 0o755)
	require.NoError(t, err)

	path := filepath.Join(dir, FileName)
	err = os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)
}

const validProjectTemplate = `
name: go-cli
type: project
version: "1.0.0"
description: "Go CLI project"
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

func TestLoader_Load(t *testing.T) {
	base := t.TempDir()
	loader, err := NewLoader(base)
	require.NoError(t, err)

	t.Run("load from relative directory", func(t *testing.T) {
		dir := filepath.Join(base, "projects", "go-cli")
		writeTemplate(t, dir, validProjectTemplate)

		tmpl, err := loader.Load("projects/go-cli")
		require.NoError(t, err)
		require.Equal(t, "go-cli", tmpl.Name)
	})

	t.Run("load from absolute directory", func(t *testing.T) {
		dir := filepath.Join(base, "abs")
		writeTemplate(t, dir, validProjectTemplate)

		tmpl, err := loader.Load(dir)
		require.NoError(t, err)
		require.Equal(t, "go-cli", tmpl.Name)
	})

	t.Run("load from absolute template.yaml path", func(t *testing.T) {
		dir := filepath.Join(base, "direct")
		writeTemplate(t, dir, validProjectTemplate)

		path := filepath.Join(dir, FileName)
		tmpl, err := loader.Load(path)
		require.NoError(t, err)
		require.Equal(t, "go-cli", tmpl.Name)
	})

	t.Run("invalid template fails validation", func(t *testing.T) {
		dir := filepath.Join(base, "invalid")
		writeTemplate(t, dir, invalidTemplate)

		_, err := loader.Load(dir)
		require.Error(t, err)
	})
}

func TestLoader_Exists(t *testing.T) {
	base := t.TempDir()
	loader, err := NewLoader(base)
	require.NoError(t, err)

	dir := filepath.Join(base, "exists")
	writeTemplate(t, dir, validProjectTemplate)

	require.True(t, loader.Exists("exists"))
	require.False(t, loader.Exists("missing"))
}

func TestLoader_Discover(t *testing.T) {
	base := t.TempDir()
	loader, err := NewLoader(base)
	require.NoError(t, err)

	writeTemplate(t, filepath.Join(base, "projects", "go-cli"), validProjectTemplate)
	writeTemplate(t, filepath.Join(base, "features", "testing"), validFeatureTemplate)
	writeTemplate(t, filepath.Join(base, "broken"), invalidTemplate)

	templates, err := loader.Discover()
	require.NoError(t, err)

	require.Len(t, templates, 2)
	require.Equal(t, "go-cli", templates["projects/go-cli"])
	require.Equal(t, "testing", templates["features/testing"])
}

func TestLoader_DiscoverByType(t *testing.T) {
	base := t.TempDir()
	loader, err := NewLoader(base)
	require.NoError(t, err)

	writeTemplate(t, filepath.Join(base, "projects", "go-cli"), validProjectTemplate)
	writeTemplate(t, filepath.Join(base, "features", "testing"), validFeatureTemplate)

	projects, err := loader.DiscoverByType(TypeProject)
	require.NoError(t, err)
	require.Len(t, projects, 1)
	require.Equal(t, "go-cli", projects["projects/go-cli"])

	features, err := loader.DiscoverByType(TypeFeature)
	require.NoError(t, err)
	require.Len(t, features, 1)
	require.Equal(t, "testing", features["features/testing"])
}

func TestLoader_GetBaseDir(t *testing.T) {
	base := t.TempDir()
	loader, err := NewLoader(base)
	require.NoError(t, err)

	require.Equal(t, base, loader.GetBaseDir())
}
