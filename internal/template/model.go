package template

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

// Template represents a complete template definition
type Template struct {
	Name         string     `yaml:"name" validate:"required"`
	Type         Type       `yaml:"type" validate:"required,oneof=project feature component"`
	Version      string     `yaml:"version" validate:"required"`
	Description  string     `yaml:"description"`
	Variables    []Variable `yaml:"variables,omitempty" validate:"dive"`
	Includes     []Include  `yaml:"includes,omitempty" validate:"dive"`
	Dependencies []string   `yaml:"dependencies,omitempty"`
	Files        []File     `yaml:"files,omitempty" validate:"dive"`
	PostInit     []PostInit `yaml:"post_init,omitempty" validate:"dive"`
}

// Variable represents a user-configurable variable with an interactive prompt
type Variable struct {
	Name    string       `yaml:"name" validate:"required"`
	Prompt  string       `yaml:"prompt"`
	Type    VariableType `yaml:"type" validate:"required,oneof=string int bool select multiselect"`
	Default any          `yaml:"default,omitempty"`
	Options []string     `yaml:"options,omitempty" validate:"required_if=Type select,required_if=Type multiselect"`
}

// Include represents another template to compose into this one
type Include struct {
	Template         string `yaml:"template" validate:"required"`
	EnabledByDefault bool   `yaml:"enabled_by_default"`
}

// File represents a template file to be rendered and written
type File struct {
	// Src is resolved relative to the directory containing template.yaml when loaded.
	Src  string `yaml:"src" validate:"required"`
	Dest string `yaml:"dest" validate:"required"`
}

// PostInit represents a command to run after scaffolding
type PostInit struct {
	Command string `yaml:"command" validate:"required"`
	WorkDir string `yaml:"workdir,omitempty"`
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
