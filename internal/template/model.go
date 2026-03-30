package template

import (
	"fmt"
	"io/fs"
)

// Type represents the semantic type of a template
type Type string

const (
	TypeProject   Type = "project"
	TypeFeature   Type = "feature"
	TypeComponent Type = "component"
)

// VariableType represents the type of input expected for a variable
type VariableType string

const (
	VariableTypeString      VariableType = "string"
	VariableTypeInt         VariableType = "int"
	VariableTypeBool        VariableType = "bool"
	VariableTypeSelect      VariableType = "select"
	VariableTypeMultiSelect VariableType = "multiselect"
)

// VariableRole represents the semantic role of a variable.
type VariableRole string

const (
	// RoleProjectName is the role for the project name variable.
	RoleProjectName VariableRole = "project_name"
)

// Template represents a complete template definition
type Template struct {
	Name         string     `yaml:"name" validate:"required"`
	Type         Type       `yaml:"type" validate:"required,oneof=project feature component"`
	Version      string     `yaml:"version" validate:"required"`
	Description  string     `yaml:"description"`
	Tags         []string   `yaml:"tags,omitempty"`
	Variables    []Variable `yaml:"variables,omitempty" validate:"dive"`
	Includes     []Include  `yaml:"includes,omitempty" validate:"dive"`
	Dependencies []string   `yaml:"dependencies,omitempty"`
	Files        []File     `yaml:"files,omitempty" validate:"dive"`
	PostInit     []PostInit `yaml:"post_init,omitempty" validate:"dive"`
}

// VariableByRole returns the variable with the given role.
func (t *Template) VariableByRole(role VariableRole) (*Variable, error) {
	for i, v := range t.Variables {
		if v.Role == role {
			return &t.Variables[i], nil
		}
	}
	return nil, fmt.Errorf("template does not have a variable with role %s", role)
}

// ProjectName returns the project name from the context.
func (t *Template) ProjectName(ctx *Context) (string, error) {
	v, err := t.VariableByRole(RoleProjectName)
	if err != nil {
		return "", err
	}

	raw, ok := ctx.Get(v.Name)
	if !ok {
		return "", fmt.Errorf("project name variable '%s' not found in context", v.Name)
	}

	name, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("project name variable '%s' must be a string", v.Name)
	}

	return name, nil
}

// RenderedFile represents a file that has been rendered but not yet written to disk.
type RenderedFile struct {
	Path    string
	Content string
}

// TemplateNode represents a resolved node in the template tree.
// It carries a guarantee that its full subtree is present and confirmed.
type TemplateNode struct {
	Template *Template
	Children []*TemplateNode
}

// ConfirmIncludes is a function that decides which optional includes should be loaded.
type ConfirmIncludes func(includes []Include) ([]Include, error)

// RenderContexts maps a template name to its specific rendering context.
type RenderContexts map[string]*Context

// Variable represents a user-configurable variable with an interactive prompt
type Variable struct {
	Name    string       `yaml:"name" validate:"required"`
	Prompt  string       `yaml:"prompt" validate:"required"`
	Type    VariableType `yaml:"type" validate:"required,oneof=string int bool select multiselect"`
	Role    VariableRole `yaml:"role,omitempty"`
	Default any          `yaml:"default,omitempty"`
	Options []string     `yaml:"options,omitempty" validate:"required_if=Type select,required_if=Type multiselect"`
}

// Include represents another template to compose into this one
type Include struct {
	Name             string `yaml:"name" validate:"required"`
	EnabledByDefault bool   `yaml:"enabled_by_default"`
}

// File represents a template file to be rendered and written
type File struct {
	// Src is resolved relative to the directory containing template.yaml when loaded.
	Src  string `yaml:"src" validate:"required"`
	Dest string `yaml:"dest" validate:"required"`
	// FS is the filesystem containing the source file. It is set during loading.
	FS fs.FS `yaml:"-"`
}

// Context holds all resolved variables for template rendering
type Context struct {
	Variables map[string]any
}

// NewTemplateContext creates a new template context with the given variables
func NewTemplateContext(vars map[string]any) *Context {
	return &Context{
		Variables: vars,
	}
}

// Get retrieves a variable value from the context
func (tc *Context) Get(key string) (any, bool) {
	val, ok := tc.Variables[key]
	return val, ok
}

// Set sets a variable value in the context
func (tc *Context) Set(key string, value any) {
	tc.Variables[key] = value
}

// Merge merges another context into this one
func (tc *Context) Merge(other *Context) {
	for k, v := range other.Variables {
		tc.Variables[k] = v
	}
}
