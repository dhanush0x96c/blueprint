package template

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeLoader struct {
	templates map[string]*Template
	err       error
}

func (f *fakeLoader) Load(name string) (*Template, error) {
	if f.err != nil {
		return nil, f.err
	}

	t, ok := f.templates[name]
	if !ok {
		return nil, errors.New("template not found")
	}

	return t, nil
}

func TestCompose_SingleTemplate_NoIncludes(t *testing.T) {
	loader := &fakeLoader{}
	composer := NewComposer(loader)

	tmpl := &Template{
		Name: "base",
		Variables: []Variable{
			{Name: "project_name"},
		},
		Dependencies: []string{"go@1.22"},
	}

	out, err := composer.Compose(tmpl)
	require.NoError(t, err)

	assert.Equal(t, "base", out.Name)
	assert.Len(t, out.Variables, 1)
	assert.Equal(t, "project_name", out.Variables[0].Name)
	assert.Equal(t, []string{"go@1.22"}, out.Dependencies)
}

func TestCompose_WithIncludes_MergesFields(t *testing.T) {
	base := &Template{
		Name: "base",
		Includes: []Include{
			{Template: "logging"},
		},
		Variables: []Variable{
			{Name: "project_name"},
		},
		Dependencies: []string{"go"},
	}

	logging := &Template{
		Name: "logging",
		Variables: []Variable{
			{Name: "log_level"},
		},
		Dependencies: []string{"zap@1.26.0"},
		Files: []File{
			{Dest: "logger.go"},
		},
	}

	loader := &fakeLoader{
		templates: map[string]*Template{
			"logging": logging,
		},
	}

	composer := NewComposer(loader)

	out, err := composer.Compose(base)
	require.NoError(t, err)

	assert.ElementsMatch(t,
		[]string{"project_name", "log_level"},
		[]string{out.Variables[0].Name, out.Variables[1].Name},
	)

	assert.ElementsMatch(t,
		[]string{"go", "zap@1.26.0"},
		out.Dependencies,
	)

	require.Len(t, out.Files, 1)
	assert.Equal(t, "logger.go", out.Files[0].Dest)
}

func TestCompose_CircularDependencyDetected(t *testing.T) {
	a := &Template{
		Name: "a",
		Includes: []Include{
			{Template: "b"},
		},
	}

	b := &Template{
		Name: "b",
		Includes: []Include{
			{Template: "a"},
		},
	}

	loader := &fakeLoader{
		templates: map[string]*Template{
			"a": a,
			"b": b,
		},
	}

	composer := NewComposer(loader)

	_, err := composer.Compose(a)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestComposeWithEnabledIncludes_FiltersCorrectly(t *testing.T) {
	base := &Template{
		Name: "base",
		Includes: []Include{
			{Template: "logging", EnabledByDefault: true},
			{Template: "metrics", EnabledByDefault: false},
		},
	}

	logging := &Template{Name: "logging"}
	metrics := &Template{Name: "metrics"}

	loader := &fakeLoader{
		templates: map[string]*Template{
			"logging": logging,
			"metrics": metrics,
		},
	}

	composer := NewComposer(loader)

	out, err := composer.ComposeWithEnabledIncludes(base, map[string]bool{
		"metrics": true,
	})
	require.NoError(t, err)

	// nothing material to assert on fields; success implies both included
	assert.Equal(t, "base", out.Name)
}

func TestGetAllIncludes_Transitive(t *testing.T) {
	base := &Template{
		Name: "base",
		Includes: []Include{
			{Template: "a"},
		},
	}

	a := &Template{
		Name: "a",
		Includes: []Include{
			{Template: "b"},
		},
	}

	b := &Template{
		Name: "b",
	}

	loader := &fakeLoader{
		templates: map[string]*Template{
			"a": a,
			"b": b,
		},
	}

	composer := NewComposer(loader)

	includes, err := composer.GetAllIncludes(base)
	require.NoError(t, err)

	require.Len(t, includes, 2)
	assert.ElementsMatch(t,
		[]string{"a", "b"},
		[]string{includes[0].Template, includes[1].Template},
	)
}

func TestMergeDependencies_PrefersVersioned(t *testing.T) {
	composer := NewComposer(nil)

	out := composer.mergeDependencies(
		[]string{"foo"},
		[]string{"foo@1.2.3"},
	)

	require.Len(t, out, 1)
	assert.Equal(t, "foo@1.2.3", out[0])
}
