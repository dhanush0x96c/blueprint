package testutil

import (
	"io/fs"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// TemplateOption defines a function that modifies a templatobjectse.Template.
type TemplateOption func(*template.Template)

// NewTemplate creates a new template.Template with the given name and options.
func NewTemplate(name string, opts ...TemplateOption) *template.Template {
	tmpl := &template.Template{
		Name:    name,
		Type:    template.TypeProject,
		Version: "1.0.0",
	}
	for _, opt := range opts {
		opt(tmpl)
	}
	return tmpl
}

func WithType(t template.Type) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Type = t
	}
}

func WithVersion(v string) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Version = v
	}
}

func WithDescription(d string) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Description = d
	}
}

func WithTags(tags ...string) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Tags = tags
	}
}

func WithVariable(v template.Variable) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Variables = append(tmpl.Variables, v)
	}
}

func WithInclude(i template.Include) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Includes = append(tmpl.Includes, i)
	}
}

func WithFile(f template.File) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Files = append(tmpl.Files, f)
	}
}

func WithDependency(d string) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Dependencies = append(tmpl.Dependencies, d)
	}
}

// NodeOption defines a function that modifies a template.Node.
type NodeOption func(*template.Node)

// NewNode creates a new template.Node with the given ID, template, and options.
func NewNode(id string, tmpl *template.Template, opts ...NodeOption) *template.Node {
	node := &template.Node{
		ID:       id,
		Template: tmpl,
	}
	for _, opt := range opts {
		opt(node)
	}
	return node
}

func WithFS(fsys fs.FS) NodeOption {
	return func(node *template.Node) {
		node.FS = fsys
	}
}

func WithPath(p string) NodeOption {
	return func(node *template.Node) {
		node.Path = p
	}
}

func WithChild(child *template.Node) NodeOption {
	return func(node *template.Node) {
		node.Children = append(node.Children, child)
	}
}

func WithInherited(key, value string) NodeOption {
	return func(node *template.Node) {
		if node.Inherited == nil {
			node.Inherited = make(map[string]string)
		}
		node.Inherited[key] = value
	}
}
