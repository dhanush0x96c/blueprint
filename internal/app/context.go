package app

import (
	"io/fs"
	"os"

	"github.com/dhanush0x96c/blueprint/internal/builtin/templates"
	"github.com/dhanush0x96c/blueprint/internal/config"
)

// Context holds runtime dependencies for the application.
type Context struct {
	Config    *config.Config
	BuiltinFS fs.FS
	LocalFS   fs.FS
	Resolver  Resolver
}

// NewContext creates a new application context.
func NewContext(cfg *config.Config) *Context {
	return &Context{
		Config:    cfg,
		LocalFS:   os.DirFS(cfg.TemplatesDir),
		BuiltinFS: templates.Templates,
		Resolver: NewChainResolver(
			&ResolverLocal{},
			&ResolverBuiltin{},
		),
	}
}
