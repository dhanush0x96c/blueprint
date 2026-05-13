package template_test

import (
	"testing"

	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompose(t *testing.T) {
	t.Run("single template no includes", func(t *testing.T) {
		loader := &testutil.FakeLoader{}
		resolver := &testutil.FakeResolver{}
		composer := template.NewComposer(resolver, loader)

		tmpl := &template.Template{
			Name: "base",
			Tags: []string{"backend", "api"},
			Variables: []template.Variable{
				{Name: "project_name"},
			},
			Dependencies: []string{"go@1.22"},
		}

		loaded := &template.LoadedTemplate{
			Template: tmpl,
			FS:       nil,
			Path:     "base",
		}

		out, err := composer.Compose(loaded, func(includes []template.Include) ([]template.Include, error) {
			return nil, nil
		})
		require.NoError(t, err)

		assert.Equal(t, "base", out.Template.Name)
		assert.Len(t, out.Children, 0)
		assert.Equal(t, []string{"go@1.22"}, out.AllDependencies())
	})

	t.Run("with includes builds tree", func(t *testing.T) {
		base := &template.Template{
			Name: "base",
			Includes: []template.Include{
				{Name: "logging", EnabledByDefault: true},
			},
			Variables: []template.Variable{
				{Name: "project_name"},
			},
			Dependencies: []string{"go"},
		}

		logging := &template.Template{
			Name: "logging",
			Variables: []template.Variable{
				{Name: "log_level"},
			},
			Dependencies: []string{"zap@1.26.0"},
			Files: []template.File{
				{Dest: "logger.go"},
			},
		}

		templates := map[string]*template.Template{
			"logging": logging,
		}

		loader := &testutil.FakeLoader{
			Templates: templates,
		}
		resolver := &testutil.FakeResolver{
			Templates: templates,
		}

		composer := template.NewComposer(resolver, loader)

		loaded := &template.LoadedTemplate{
			Template: base,
			FS:       nil,
			Path:     "base",
		}

		out, err := composer.Compose(loaded, func(includes []template.Include) ([]template.Include, error) {
			return includes, nil
		})
		require.NoError(t, err)

		assert.Equal(t, "base", out.Template.Name)
		require.Len(t, out.Children, 1)
		assert.Equal(t, "logging", out.Children[0].Template.Name)

		assert.ElementsMatch(t,
			[]string{"go", "zap@1.26.0"},
			out.AllDependencies(),
		)
	})

	t.Run("circular dependency detected", func(t *testing.T) {
		a := &template.Template{
			Name: "a",
			Includes: []template.Include{
				{Name: "b", EnabledByDefault: true},
			},
		}

		b := &template.Template{
			Name: "b",
			Includes: []template.Include{
				{Name: "a", EnabledByDefault: true},
			},
		}

		templates := map[string]*template.Template{
			"a": a,
			"b": b,
		}

		loader := &testutil.FakeLoader{
			Templates: templates,
		}
		resolver := &testutil.FakeResolver{
			Templates: templates,
		}

		composer := template.NewComposer(resolver, loader)

		loaded := &template.LoadedTemplate{
			Template: a,
			FS:       nil,
			Path:     "a",
		}

		_, err := composer.Compose(loaded, func(includes []template.Include) ([]template.Include, error) {
			return includes, nil
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency")
	})

	t.Run("optional includes confirm called", func(t *testing.T) {
		base := &template.Template{
			Name: "base",
			Includes: []template.Include{
				{Name: "logging", EnabledByDefault: false},
				{Name: "metrics", EnabledByDefault: false},
			},
		}

		logging := &template.Template{
			Name: "logging",
		}
		metrics := &template.Template{
			Name: "metrics",
		}

		templates := map[string]*template.Template{
			"logging": logging,
			"metrics": metrics,
		}

		loader := &testutil.FakeLoader{
			Templates: templates,
		}
		resolver := &testutil.FakeResolver{
			Templates: templates,
		}

		composer := template.NewComposer(resolver, loader)

		loaded := &template.LoadedTemplate{
			Template: base,
			FS:       nil,
			Path:     "base",
		}

		// Enable only logging
		confirm := func(includes []template.Include) ([]template.Include, error) {
			var enabled []template.Include
			for _, inc := range includes {
				if inc.Name == "logging" {
					enabled = append(enabled, inc)
				}
			}
			return enabled, nil
		}

		out, err := composer.Compose(loaded, confirm)
		require.NoError(t, err)

		assert.Equal(t, "base", out.Template.Name)
		require.Len(t, out.Children, 1)
		assert.Equal(t, "logging", out.Children[0].Template.Name)
	})

	t.Run("assigns IDs", func(t *testing.T) {
		root := &template.Template{
			Name: "root",
			Includes: []template.Include{
				{Name: "child0", EnabledByDefault: true},
				{Name: "child1", EnabledByDefault: true},
			},
		}

		child0 := &template.Template{
			Name: "child0",
		}

		child1 := &template.Template{
			Name: "child1",
			Includes: []template.Include{
				{Name: "grandchild0", EnabledByDefault: true},
				{Name: "grandchild1", EnabledByDefault: true},
			},
		}

		grandchild0 := &template.Template{
			Name: "grandchild0",
		}

		grandchild1 := &template.Template{
			Name: "grandchild1",
		}

		templates := map[string]*template.Template{
			"child0":      child0,
			"child1":      child1,
			"grandchild0": grandchild0,
			"grandchild1": grandchild1,
		}

		loader := &testutil.FakeLoader{
			Templates: templates,
		}
		resolver := &testutil.FakeResolver{
			Templates: templates,
		}

		composer := template.NewComposer(resolver, loader)

		loaded := &template.LoadedTemplate{
			Template: root,
			FS:       nil,
			Path:     "root",
		}

		out, err := composer.Compose(loaded, func(includes []template.Include) ([]template.Include, error) {
			return includes, nil
		})
		require.NoError(t, err)

		assert.Equal(t, "0", out.ID)
		require.Len(t, out.Children, 2)

		assert.Equal(t, "0.0", out.Children[0].ID)
		assert.Equal(t, "child0", out.Children[0].Template.Name)

		assert.Equal(t, "0.1", out.Children[1].ID)
		assert.Equal(t, "child1", out.Children[1].Template.Name)

		require.Len(t, out.Children[1].Children, 2)
		assert.Equal(t, "0.1.0", out.Children[1].Children[0].ID)
		assert.Equal(t, "grandchild0", out.Children[1].Children[0].Template.Name)
		assert.Equal(t, "0.1.1", out.Children[1].Children[1].ID)
		assert.Equal(t, "grandchild1", out.Children[1].Children[1].Template.Name)
	})
}
