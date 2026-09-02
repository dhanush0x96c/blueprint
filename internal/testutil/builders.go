package testutil

import (
	"io/fs"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// TemplateOption defines a function that modifies a template.Template.
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

// WithType sets the template type.
func WithType(t template.Type) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Type = t
	}
}

// WithVersion sets the template version.
func WithVersion(v string) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Version = v
	}
}

// WithDescription sets the template description.
func WithDescription(d string) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Description = d
	}
}

// WithTags sets the template tags.
func WithTags(tags ...string) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Tags = tags
	}
}

// WithVariable appends a variable to the template.
func WithVariable(v template.Variable) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Variables = append(tmpl.Variables, v)
	}
}

// WithInclude appends an include to the template.
func WithInclude(i template.Include) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Includes = append(tmpl.Includes, i)
	}
}

// WithFile appends a file to the template.
func WithFile(f template.File) TemplateOption {
	return func(tmpl *template.Template) {
		tmpl.Files = append(tmpl.Files, f)
	}
}

// WithDependency appends a dependency to the template.
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

// WithFS sets the filesystem on a template node.
func WithFS(fsys fs.FS) NodeOption {
	return func(node *template.Node) {
		node.FS = fsys
	}
}

// WithPath sets the path on a template node.
func WithPath(p string) NodeOption {
	return func(node *template.Node) {
		node.Path = p
	}
}

// WithChild appends a child node to a template node.
func WithChild(child *template.Node) NodeOption {
	return func(node *template.Node) {
		node.Children = append(node.Children, child)
	}
}

// WithInherited sets an inherited variable mapping on a template node.
func WithInherited(key, value string) NodeOption {
	return func(node *template.Node) {
		if node.Inherited == nil {
			node.Inherited = make(map[string]string)
		}
		node.Inherited[key] = value
	}
}
