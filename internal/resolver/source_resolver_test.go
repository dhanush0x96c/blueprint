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

func TestSourceResolver_Exists(t *testing.T) {
	base := t.TempDir()
	r := NewSourceResolver(Source{
		Name:       "test",
		Type:       SourceTypeUser,
		Filesystem: os.DirFS(base),
	})

	templatePath := "exists"
	dir := filepath.Join(base, templatePath)
	writeTemplate(t, dir, validProjectTemplate)

	require.True(t, r.Exists("go-cli"))
	require.False(t, r.Exists("exists"))
	require.False(t, r.Exists("missing"))
}

const validTemplateWithTags = `
name: go-api
type: project
version: "1.0.0"
description: "Go API project"
tags: ["go", "api"]
variables:
  - name: app_name
    prompt: "App name?"
    type: string
    role: project_name
`

const validFeatureTemplateWithTags = `
name: auth
type: feature
version: "1.0.0"
description: "Authentication"
tags: ["auth", "security"]
`

func TestSourceResolver_Discover(t *testing.T) {
	base := t.TempDir()
	r := NewSourceResolver(Source{
		Name:       "test",
		Type:       SourceTypeUser,
		Filesystem: os.DirFS(base),
	})

	writeTemplate(t, filepath.Join(base, "projects", "go-cli"), validProjectTemplate)
	writeTemplate(t, filepath.Join(base, "projects", "go-api"), validTemplateWithTags)
	writeTemplate(t, filepath.Join(base, "features", "testing"), validFeatureTemplate)
	writeTemplate(t, filepath.Join(base, "features", "auth"), validFeatureTemplateWithTags)
	writeTemplate(t, filepath.Join(base, "broken"), invalidTemplate)

	t.Run("all templates", func(t *testing.T) {
		templates, err := r.Discover(template.DiscoverOptions{IgnoreErrors: true})
		require.NoError(t, err)
		require.Len(t, templates, 4)
	})

	t.Run("filter by type", func(t *testing.T) {
		templates, err := r.Discover(template.DiscoverOptions{
			Type:         template.TypeProject,
			IgnoreErrors: true,
		})
		require.NoError(t, err)
		require.Len(t, templates, 2)
		for _, tmpl := range templates {
			require.Equal(t, template.TypeProject, tmpl.Type)
		}
	})

	t.Run("filter by tag", func(t *testing.T) {
		templates, err := r.Discover(template.DiscoverOptions{
			Tags:         []string{"go"},
			IgnoreErrors: true,
		})
		require.NoError(t, err)
		require.Len(t, templates, 1)
		require.Equal(t, "go-api", templates["projects/go-api"].Name)
	})

	t.Run("filter by multiple tags", func(t *testing.T) {
		templates, err := r.Discover(template.DiscoverOptions{
			Tags:         []string{"go", "auth"},
			IgnoreErrors: true,
		})
		require.NoError(t, err)
		require.Len(t, templates, 2)
		require.Contains(t, templates, "projects/go-api")
		require.Contains(t, templates, "features/auth")
	})

	t.Run("filter by type and tag", func(t *testing.T) {
		templates, err := r.Discover(template.DiscoverOptions{
			Type:         template.TypeFeature,
			Tags:         []string{"auth"},
			IgnoreErrors: true,
		})
		require.NoError(t, err)
		require.Len(t, templates, 1)
		require.Equal(t, "auth", templates["features/auth"].Name)
	})

	t.Run("error on invalid template when IgnoreErrors is false", func(t *testing.T) {
		_, err := r.Discover(template.DiscoverOptions{IgnoreErrors: false})
		require.Error(t, err)
	})
}
