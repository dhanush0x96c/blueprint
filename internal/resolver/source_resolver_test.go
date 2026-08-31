// Package resolver_test contains unit tests for the resolver package.
package resolver_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dhanush0x96c/blueprint/internal/resolver"
	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestSourceResolver_Exists(t *testing.T) {
	base := t.TempDir()
	r := resolver.NewSourceResolver(resolver.Source{
		Name:       "test",
		Type:       resolver.SourceTypeUser,
		Filesystem: os.DirFS(base),
	})

	// Directory name differs from template metadata name.
	// Exists() is name-based, not path-based.
	dir := filepath.Join(base, "exists")
	testutil.WriteTemplateStruct(t, dir, testutil.NewTemplate("go-cli",
		testutil.WithVariable(template.Variable{Name: "app", Role: template.RoleProjectName, Prompt: "?", Type: template.VariableTypeString})))

	t.Run("returns true when template metadata name exists", func(t *testing.T) {
		require.True(t, r.Exists("go-cli"))
	})

	t.Run("returns false for directory path or directory name lookups", func(t *testing.T) {
		require.False(t, r.Exists("exists"))
		require.False(t, r.Exists("projects/go-cli"))
	})

	t.Run("returns false when template name does not exist", func(t *testing.T) {
		require.False(t, r.Exists("missing"))
	})
}

func TestSourceResolver_Discover(t *testing.T) {
	base := t.TempDir()
	r := resolver.NewSourceResolver(resolver.Source{
		Name:       "test",
		Type:       resolver.SourceTypeUser,
		Filesystem: os.DirFS(base),
	})

	testutil.WriteTemplateStruct(t, filepath.Join(base, "projects", "go-cli"), testutil.NewTemplate("go-cli",
		testutil.WithVariable(template.Variable{Name: "app", Role: template.RoleProjectName, Prompt: "?", Type: template.VariableTypeString})))
	testutil.WriteTemplateStruct(t, filepath.Join(base, "projects", "go-api"), testutil.NewTemplate("go-api",
		testutil.WithVariable(template.Variable{Name: "app", Role: template.RoleProjectName, Prompt: "?", Type: template.VariableTypeString}),
		testutil.WithTags("go", "api")))
	testutil.WriteTemplateStruct(t, filepath.Join(base, "features", "testing"), testutil.NewTemplate("testing",
		testutil.WithType(template.TypeFeature)))
	testutil.WriteTemplateStruct(t, filepath.Join(base, "features", "auth"), testutil.NewTemplate("auth",
		testutil.WithType(template.TypeFeature),
		testutil.WithTags("auth", "security")))
	testutil.WriteTemplateStruct(t, filepath.Join(base, "broken"), &template.Template{Name: "invalid", Type: template.TypeProject})

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
