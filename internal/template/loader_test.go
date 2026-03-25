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
	loader := NewLoader(os.DirFS(base))

	t.Run("load from relative directory", func(t *testing.T) {
		dir := filepath.Join(base, "projects", "go-cli")
		writeTemplate(t, dir, validProjectTemplate)

		tmpl, err := loader.Load("projects/go-cli")
		require.NoError(t, err)
		require.Equal(t, "go-cli", tmpl.Name)
	})

	t.Run("load from template.yaml path", func(t *testing.T) {
		dir := filepath.Join(base, "direct")
		writeTemplate(t, dir, validProjectTemplate)

		path := filepath.Join("direct", FileName)
		tmpl, err := loader.Load(path)
		require.NoError(t, err)
		require.Equal(t, "go-cli", tmpl.Name)
	})

	t.Run("invalid template fails validation", func(t *testing.T) {
		templateName := "invalid"
		dir := filepath.Join(base, templateName)
		writeTemplate(t, dir, invalidTemplate)

		_, err := loader.Load(templateName)
		require.Error(t, err)
	})
}

func TestLoader_Exists(t *testing.T) {
	base := t.TempDir()
	loader := NewLoader(os.DirFS(base))

	templateName := "exists"
	dir := filepath.Join(base, templateName)
	writeTemplate(t, dir, validProjectTemplate)

	require.True(t, loader.Exists(templateName))
	require.False(t, loader.Exists("missing"))
}

func TestLoader_Discover(t *testing.T) {
	base := t.TempDir()
	loader := NewLoader(os.DirFS(base))

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
	loader := NewLoader(os.DirFS(base))

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

func TestLoader_LoadTags(t *testing.T) {
	base := t.TempDir()
	loader := NewLoader(os.DirFS(base))

	const templateWithTags = `
name: tagged-template
type: project
version: "1.0.0"
description: "Template with tags"
tags: ["go", "cli", "testing"]
`

	const templateWithoutTags = `
name: no-tags
type: feature
version: "1.0.0"
description: "Template without tags"
`

	t.Run("loads tags when present", func(t *testing.T) {
		dir := filepath.Join(base, "with-tags")
		writeTemplate(t, dir, templateWithTags)

		tmpl, err := loader.Load("with-tags")
		require.NoError(t, err)
		require.Equal(t, "tagged-template", tmpl.Name)
		require.Len(t, tmpl.Tags, 3)
		require.Equal(t, []string{"go", "cli", "testing"}, tmpl.Tags)
	})

	t.Run("handles missing tags", func(t *testing.T) {
		dir := filepath.Join(base, "without-tags")
		writeTemplate(t, dir, templateWithoutTags)

		tmpl, err := loader.Load("without-tags")
		require.NoError(t, err)
		require.Equal(t, "no-tags", tmpl.Name)
		require.Nil(t, tmpl.Tags)
	})
}

