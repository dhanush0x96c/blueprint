package testutil

import (
	"errors"
	"io/fs"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

type FakeResolver struct {
	Templates map[string]*template.Template
}

func (f *FakeResolver) Resolve(ref template.TemplateRef) (*template.ResolvedTemplate, error) {
	if _, ok := f.Templates[ref.Name]; !ok {
		return nil, errors.New("template not found")
	}
	return &template.ResolvedTemplate{
		Path: ref.Name,
		FS:   nil,
	}, nil
}

type FakeLoader struct {
	Templates map[string]*template.Template
	Err       error
}

func (f *FakeLoader) Load(fsys fs.FS, pth string) (*template.LoadedTemplate, error) {
	if f.Err != nil {
		return nil, f.Err
	}

	tmpl, ok := f.Templates[pth]
	if !ok {
		return nil, errors.New("template not found")
	}

	return &template.LoadedTemplate{
		Template: tmpl,
		FS:       fsys,
		Path:     pth,
	}, nil
}
