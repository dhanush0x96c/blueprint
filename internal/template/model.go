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
	Name         string     `yaml:"name"`
	Type         Type       `yaml:"type"`
	Version      string     `yaml:"version"`
	Description  string     `yaml:"description"`
	Variables    []Variable `yaml:"variables,omitempty"`
	Includes     []Include  `yaml:"includes,omitempty"`
	Dependencies []string   `yaml:"dependencies,omitempty"`
	Files        []File     `yaml:"files,omitempty"`
	PostInit     []PostInit `yaml:"post_init,omitempty"`
}

// Variable represents a user-configurable variable with an interactive prompt
type Variable struct {
	Name    string       `yaml:"name"`
	Prompt  string       `yaml:"prompt"`
	Type    VariableType `yaml:"type"`
	Default any          `yaml:"default,omitempty"`
	Options []string     `yaml:"options,omitempty"`
}

// Include represents another template to compose into this one
type Include struct {
	Template         string `yaml:"template"`
	EnabledByDefault bool   `yaml:"enabled_by_default"`
}

// File represents a template file to be rendered and written
type File struct {
	Src  string `yaml:"src"`
	Dest string `yaml:"dest"`
}

// PostInit represents a command to run after scaffolding
type PostInit struct {
	Command string `yaml:"command"`
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
