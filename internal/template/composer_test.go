package template

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeResolver struct {
	templates map[string]*Template
}

func (f *fakeResolver) Resolve(ref TemplateRef) (*ResolvedTemplate, error) {
	if _, ok := f.templates[ref.Name]; !ok {
		return nil, errors.New("template not found")
	}
	return &ResolvedTemplate{
		Path: ref.Name,
		FS:   nil, // Not used in fakeLoader
	}, nil
}

type fakeLoader struct {
	templates map[string]*Template
	err       error
}

func (f *fakeLoader) Load(fsys fs.FS, pth string) (*Template, error) {
	if f.err != nil {
		return nil, f.err
	}

	t, ok := f.templates[pth]
	if !ok {
		return nil, errors.New("template not found")
	}

	return t, nil
}

func TestCompose_SingleTemplate_NoIncludes(t *testing.T) {
	loader := &fakeLoader{}
	resolver := &fakeResolver{}
	composer := NewComposer(resolver, loader)

	tmpl := &Template{
		Name: "base",
		Tags: []string{"backend", "api"},
		Variables: []Variable{
			{Name: "project_name"},
		},
		Dependencies: []string{"go@1.22"},
	}

	out, err := composer.Compose(tmpl)
	require.NoError(t, err)

	assert.Equal(t, "base", out.Name)
	assert.Equal(t, []string{"backend", "api"}, out.Tags)
	assert.Len(t, out.Variables, 1)
	assert.Equal(t, "project_name", out.Variables[0].Name)
	assert.Equal(t, []string{"go@1.22"}, out.Dependencies)
}

func TestCompose_WithIncludes_MergesFields(t *testing.T) {
	base := &Template{
		Name: "base",
		Includes: []Include{
			{Name: "logging"},
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

	templates := map[string]*Template{
		"logging": logging,
	}

	loader := &fakeLoader{
		templates: templates,
	}
	resolver := &fakeResolver{
		templates: templates,
	}

	composer := NewComposer(resolver, loader)

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
			{Name: "b"},
		},
	}

	b := &Template{
		Name: "b",
		Includes: []Include{
			{Name: "a"},
		},
	}

	templates := map[string]*Template{
		"a": a,
		"b": b,
	}

	loader := &fakeLoader{
		templates: templates,
	}
	resolver := &fakeResolver{
		templates: templates,
	}

	composer := NewComposer(resolver, loader)

	_, err := composer.Compose(a)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestComposeWithEnabledIncludes_FiltersCorrectly(t *testing.T) {
	base := &Template{
		Name: "base",
		Tags: []string{"service"},
		Includes: []Include{
			{Name: "logging", EnabledByDefault: true},
			{Name: "metrics", EnabledByDefault: false},
		},
	}

	logging := &Template{Name: "logging"}
	metrics := &Template{Name: "metrics"}

	templates := map[string]*Template{
		"logging": logging,
		"metrics": metrics,
	}

	loader := &fakeLoader{
		templates: templates,
	}
	resolver := &fakeResolver{
		templates: templates,
	}

	composer := NewComposer(resolver, loader)

	out, err := composer.ComposeWithEnabledIncludes(base, map[string]bool{
		"metrics": true,
	})
	require.NoError(t, err)

	assert.Equal(t, "base", out.Name)
	assert.Equal(t, []string{"service"}, out.Tags)
}

func TestGetAllIncludes_Transitive(t *testing.T) {
	base := &Template{
		Name: "base",
		Includes: []Include{
			{Name: "a"},
		},
	}

	a := &Template{
		Name: "a",
		Includes: []Include{
			{Name: "b"},
		},
	}

	b := &Template{
		Name: "b",
	}

	templates := map[string]*Template{
		"a": a,
		"b": b,
	}

	loader := &fakeLoader{
		templates: templates,
	}
	resolver := &fakeResolver{
		templates: templates,
	}

	composer := NewComposer(resolver, loader)

	includes, err := composer.GetAllIncludes(base)
	require.NoError(t, err)

	require.Len(t, includes, 2)
	assert.ElementsMatch(t,
		[]string{"a", "b"},
		[]string{includes[0].Name, includes[1].Name},
	)
}

func TestMergeDependencies_PrefersVersioned(t *testing.T) {
	composer := NewComposer(nil, nil)

	out := composer.mergeDependencies(
		[]string{"foo"},
		[]string{"foo@1.2.3"},
	)

	require.Len(t, out, 1)
	assert.Equal(t, "foo@1.2.3", out[0])
}
