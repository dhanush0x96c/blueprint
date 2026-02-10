package app

import (
	"io/fs"
	"path"
)

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

type ResolverBuiltin struct{}

func (r *ResolverBuiltin) Resolve(ctx *Context, ref TemplateRef) (*ResolvedTemplate, error) {
	return ResolveFromFS(ctx.BuiltinFS, ref)
}

type ResolverLocal struct{}

func (r *ResolverLocal) Resolve(ctx *Context, ref TemplateRef) (*ResolvedTemplate, error) {
	return ResolveFromFS(ctx.LocalFS, ref)
}
