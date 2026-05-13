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

	templatePath := "exists"
	dir := filepath.Join(base, templatePath)
	testutil.WriteTemplate(t, dir, testutil.ValidProjectTemplate)

	require.True(t, r.Exists("go-cli"))
	require.False(t, r.Exists("exists"))
	require.False(t, r.Exists("missing"))
}

func TestSourceResolver_Discover(t *testing.T) {
	base := t.TempDir()
	r := resolver.NewSourceResolver(resolver.Source{
		Name:       "test",
		Type:       resolver.SourceTypeUser,
		Filesystem: os.DirFS(base),
	})

	testutil.WriteTemplate(t, filepath.Join(base, "projects", "go-cli"), testutil.ValidProjectTemplate)
	testutil.WriteTemplate(t, filepath.Join(base, "projects", "go-api"), testutil.ValidTemplateWithTags)
	testutil.WriteTemplate(t, filepath.Join(base, "features", "testing"), testutil.ValidFeatureTemplate)
	testutil.WriteTemplate(t, filepath.Join(base, "features", "auth"), testutil.ValidFeatureTemplateWithTags)
	testutil.WriteTemplate(t, filepath.Join(base, "broken"), testutil.InvalidTemplate)

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
