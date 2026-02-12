package app

import (
	"io/fs"
	"path"
)

// ResolveFromFS resolves a template from a given file system.
func ResolveFromFS(rootFS fs.FS, ref TemplateRef) (*ResolvedTemplate, error) {
	templatePath := path.Join(ref.Type.Folder(), ref.Name)

	_, err := fs.Stat(rootFS, templatePath)
	if err != nil {
		return nil, ErrTemplateNotFound
	}

	return &ResolvedTemplate{
		FS:   rootFS,
		Path: templatePath,
	}, nil
}

// ResolverBuiltin resolves templates from the builtin (embedded) file system.
type ResolverBuiltin struct{}

// Resolve resolves templates from the builtin (embedded) file system.
func (r *ResolverBuiltin) Resolve(ctx *Context, ref TemplateRef) (*ResolvedTemplate, error) {
	return ResolveFromFS(ctx.BuiltinFS, ref)
}

// ResolverLocal resolves user defined templates from the local file system.
type ResolverLocal struct{}

// Resolve resolves user defined templates from the local file system.
func (r *ResolverLocal) Resolve(ctx *Context, ref TemplateRef) (*ResolvedTemplate, error) {
	return ResolveFromFS(ctx.LocalFS, ref)
}
