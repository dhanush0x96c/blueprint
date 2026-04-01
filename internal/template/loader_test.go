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

func TestLoader_Load(t *testing.T) {
	base := t.TempDir()
	fsys := os.DirFS(base)
	loader := NewLoader()

	t.Run("load from relative directory", func(t *testing.T) {
		dir := filepath.Join(base, "projects", "go-cli")
		writeTemplate(t, dir, validProjectTemplate)

		tmpl, err := loader.Load(fsys, "projects/go-cli")
		require.NoError(t, err)
		require.Equal(t, "go-cli", tmpl.Template.Name)
	})

	t.Run("load from template.yaml path", func(t *testing.T) {
		dir := filepath.Join(base, "direct")
		writeTemplate(t, dir, validProjectTemplate)

		path := filepath.Join("direct", FileName)
		tmpl, err := loader.Load(fsys, path)
		require.NoError(t, err)
		require.Equal(t, "go-cli", tmpl.Template.Name)
	})

	t.Run("invalid template fails validation", func(t *testing.T) {
		templateName := "invalid"
		dir := filepath.Join(base, templateName)
		writeTemplate(t, dir, invalidTemplate)

		_, err := loader.Load(fsys, templateName)
		require.Error(t, err)
	})
}

func TestLoader_LoadTags(t *testing.T) {
	base := t.TempDir()
	fsys := os.DirFS(base)
	loader := NewLoader()

	const templateWithTags = `
name: tagged-template
type: project
version: "1.0.0"
description: "Template with tags"
tags: ["go", "cli", "testing"]
variables:
  - name: app_name
    prompt: "App name?"
    type: string
    role: project_name
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

		tmpl, err := loader.Load(fsys, "with-tags")
		require.NoError(t, err)
		require.Equal(t, "tagged-template", tmpl.Template.Name)
		require.Len(t, tmpl.Template.Tags, 3)
		require.Equal(t, []string{"go", "cli", "testing"}, tmpl.Template.Tags)
	})

	t.Run("handles missing tags", func(t *testing.T) {
		dir := filepath.Join(base, "without-tags")
		writeTemplate(t, dir, templateWithoutTags)

		tmpl, err := loader.Load(fsys, "without-tags")
		require.NoError(t, err)
		require.Equal(t, "no-tags", tmpl.Template.Name)
		require.Nil(t, tmpl.Template.Tags)
	})
}
