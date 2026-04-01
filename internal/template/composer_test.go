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

func (f *fakeLoader) Load(fsys fs.FS, pth string) (*LoadedTemplate, error) {
	if f.err != nil {
		return nil, f.err
	}

	t, ok := f.templates[pth]
	if !ok {
		return nil, errors.New("template not found")
	}

	return &LoadedTemplate{
		Template: t,
		FS:       fsys,
		Path:     pth,
	}, nil
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

	loaded := &LoadedTemplate{
		Template: tmpl,
		FS:       nil,
		Path:     "base",
	}

	out, err := composer.Compose(loaded, func(includes []Include) ([]Include, error) {
		return nil, nil
	})
	require.NoError(t, err)

	assert.Equal(t, "base", out.Template.Name)
	assert.Len(t, out.Children, 0)
	assert.Equal(t, []string{"go@1.22"}, out.AllDependencies())
}

func TestCompose_WithIncludes_BuildsTree(t *testing.T) {
	base := &Template{
		Name: "base",
		Includes: []Include{
			{Name: "logging", EnabledByDefault: true},
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

	loaded := &LoadedTemplate{
		Template: base,
		FS:       nil,
		Path:     "base",
	}

	out, err := composer.Compose(loaded, func(includes []Include) ([]Include, error) {
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
}

func TestCompose_CircularDependencyDetected(t *testing.T) {
	a := &Template{
		Name: "a",
		Includes: []Include{
			{Name: "b", EnabledByDefault: true},
		},
	}

	b := &Template{
		Name: "b",
		Includes: []Include{
			{Name: "a", EnabledByDefault: true},
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

	loaded := &LoadedTemplate{
		Template: a,
		FS:       nil,
		Path:     "a",
	}

	_, err := composer.Compose(loaded, func(includes []Include) ([]Include, error) {
		return includes, nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestCompose_OptionalIncludes_ConfirmCalled(t *testing.T) {
	base := &Template{
		Name: "base",
		Includes: []Include{
			{Name: "logging", EnabledByDefault: false},
			{Name: "metrics", EnabledByDefault: false},
		},
	}

	logging := &Template{
		Name: "logging",
	}
	metrics := &Template{
		Name: "metrics",
	}

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

	loaded := &LoadedTemplate{
		Template: base,
		FS:       nil,
		Path:     "base",
	}

	// Enable only logging
	confirm := func(includes []Include) ([]Include, error) {
		var enabled []Include
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
}
