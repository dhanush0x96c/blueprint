package testutil

import (
	"errors"
	"io/fs"

	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/template/loader"
)

// FakeResolver is a test fake for template resolution.
type FakeResolver struct {
	Templates map[string]*template.Template
}

// Resolve resolves a template reference against the faked templates.
func (f *FakeResolver) Resolve(ref template.Ref) (*template.ResolvedTemplate, error) {
	if _, ok := f.Templates[ref.Name]; !ok {
		return nil, errors.New("template not found")
	}
	return &template.ResolvedTemplate{
		Path: ref.Name,
		FS:   nil,
	}, nil
}

// FakeLoader is a test fake for composer.Loader.
type FakeLoader struct {
	Templates map[string]*template.Template
	Err       error
}

// Load loads a template from the faked templates map.
func (f *FakeLoader) Load(fsys fs.FS, pth string) (*loader.LoadedTemplate, error) {
	if f.Err != nil {
		return nil, f.Err
	}

	tmpl, ok := f.Templates[pth]
	if !ok {
		return nil, errors.New("template not found")
	}

	return &loader.LoadedTemplate{
		Template: tmpl,
		FS:       fsys,
		Path:     pth,
	}, nil
}
