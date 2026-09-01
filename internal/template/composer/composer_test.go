// Package composer_test contains unit tests for the composer package.
package composer_test

import (
	"testing"

	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/template/composer"
	"github.com/dhanush0x96c/blueprint/internal/template/loader"
	"github.com/dhanush0x96c/blueprint/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompose(t *testing.T) {
	t.Run("single template no includes", func(t *testing.T) {
		l := &testutil.FakeLoader{}
		resolver := &testutil.FakeResolver{}
		c := composer.NewComposer(resolver, l)

		tmpl := testutil.NewTemplate("base",
			testutil.WithTags("backend", "api"),
			testutil.WithVariable(template.Variable{Name: "project_name"}),
			testutil.WithDependency("go@1.22"))

		loaded := &loader.LoadedTemplate{
			Template: tmpl,
			FS:       nil,
			Path:     "base",
		}

		out, err := c.Compose(loaded, func(includes []template.Include) ([]template.Include, error) {
			return nil, nil
		})
		require.NoError(t, err)

		assert.Equal(t, "base", out.Template.Name)
		assert.Len(t, out.Children, 0)
		assert.Equal(t, []string{"go@1.22"}, out.AllDependencies())
	})

	t.Run("with includes builds tree", func(t *testing.T) {
		base := testutil.NewTemplate("base",
			testutil.WithInclude(template.Include{Name: "logging", EnabledByDefault: true}),
			testutil.WithVariable(template.Variable{Name: "project_name"}),
			testutil.WithDependency("go"))

		logging := testutil.NewTemplate("logging",
			testutil.WithVariable(template.Variable{Name: "log_level"}),
			testutil.WithDependency("zap@1.26.0"),
			testutil.WithFile(template.File{Dest: "logger.go"}))

		templates := map[string]*template.Template{
			"logging": logging,
		}

		l := &testutil.FakeLoader{
			Templates: templates,
		}
		resolver := &testutil.FakeResolver{
			Templates: templates,
		}

		c := composer.NewComposer(resolver, l)

		loaded := &loader.LoadedTemplate{
			Template: base,
			FS:       nil,
			Path:     "base",
		}

		out, err := c.Compose(loaded, func(includes []template.Include) ([]template.Include, error) {
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
		a := testutil.NewTemplate("a",
			testutil.WithInclude(template.Include{Name: "b", EnabledByDefault: true}))

		b := testutil.NewTemplate("b",
			testutil.WithInclude(template.Include{Name: "a", EnabledByDefault: true}))

		templates := map[string]*template.Template{
			"a": a,
			"b": b,
		}

		l := &testutil.FakeLoader{
			Templates: templates,
		}
		resolver := &testutil.FakeResolver{
			Templates: templates,
		}

		c := composer.NewComposer(resolver, l)

		loaded := &loader.LoadedTemplate{
			Template: a,
			FS:       nil,
			Path:     "a",
		}

		_, err := c.Compose(loaded, func(includes []template.Include) ([]template.Include, error) {
			return includes, nil
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency")
	})

	t.Run("optional includes confirm called", func(t *testing.T) {
		base := testutil.NewTemplate("base",
			testutil.WithInclude(template.Include{Name: "logging", EnabledByDefault: false}),
			testutil.WithInclude(template.Include{Name: "metrics", EnabledByDefault: false}))

		logging := testutil.NewTemplate("logging")
		metrics := testutil.NewTemplate("metrics")

		templates := map[string]*template.Template{
			"logging": logging,
			"metrics": metrics,
		}

		l := &testutil.FakeLoader{
			Templates: templates,
		}
		resolver := &testutil.FakeResolver{
			Templates: templates,
		}

		c := composer.NewComposer(resolver, l)

		loaded := &loader.LoadedTemplate{
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

		out, err := c.Compose(loaded, confirm)
		require.NoError(t, err)

		assert.Equal(t, "base", out.Template.Name)
		require.Len(t, out.Children, 1)
		assert.Equal(t, "logging", out.Children[0].Template.Name)
	})

	t.Run("assigns IDs", func(t *testing.T) {
		root := testutil.NewTemplate("root",
			testutil.WithInclude(template.Include{Name: "child0", EnabledByDefault: true}),
			testutil.WithInclude(template.Include{Name: "child1", EnabledByDefault: true}))

		child0 := testutil.NewTemplate("child0")

		child1 := testutil.NewTemplate("child1",
			testutil.WithInclude(template.Include{Name: "grandchild0", EnabledByDefault: true}),
			testutil.WithInclude(template.Include{Name: "grandchild1", EnabledByDefault: true}))

		grandchild0 := testutil.NewTemplate("grandchild0")

		grandchild1 := testutil.NewTemplate("grandchild1")

		templates := map[string]*template.Template{
			"child0":      child0,
			"child1":      child1,
			"grandchild0": grandchild0,
			"grandchild1": grandchild1,
		}

		l := &testutil.FakeLoader{
			Templates: templates,
		}
		resolver := &testutil.FakeResolver{
			Templates: templates,
		}

		c := composer.NewComposer(resolver, l)

		loaded := &loader.LoadedTemplate{
			Template: root,
			FS:       nil,
			Path:     "root",
		}

		out, err := c.Compose(loaded, func(includes []template.Include) ([]template.Include, error) {
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
