package app

import (
	"io/fs"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

type TemplateRef struct {
	Name string
	Type template.Type
}

type ResolvedTemplate struct {
	FS   fs.FS
	Path string
}

type Resolver interface {
	Resolve(ctx *Context, ref TemplateRef) (*ResolvedTemplate, error)
}
