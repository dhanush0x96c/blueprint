package template_test

import (
	"testing"
	"testing/fstest"

	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_Validate(t *testing.T) {
	v := template.NewValidator()

	t.Run("valid template passes", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "test",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName},
			},
		}

		err := v.Validate(tmpl)
		require.NoError(t, err)
	})

	t.Run("missing required fields fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name: "", // missing
			Type: template.TypeProject,
			// Version missing
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Name")
		assert.Contains(t, err.Error(), "Version")
	})

	t.Run("invalid type fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "test",
			Type:    "invalid",
			Version: "1.0.0",
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Type")
	})
}

func TestValidator_ValidateVariables(t *testing.T) {
	v := template.NewValidator()

	t.Run("duplicate variable names fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "test",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "app_name", Prompt: "Another?", Type: template.VariableTypeString},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate variable name")
		assert.Contains(t, err.Error(), "app_name")
	})

	t.Run("missing prompt fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "test",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app_name", Prompt: "", Type: template.VariableTypeString},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Prompt")
	})

	t.Run("select without options fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "test",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "choice", Prompt: "Choose?", Type: template.VariableTypeSelect, Options: []string{}},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "options required")
	})

	t.Run("multiselect without options fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "test",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "choices", Prompt: "Choose?", Type: template.VariableTypeMultiSelect, Options: nil},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "options required")
	})

	t.Run("select with options passes", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "test",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "choice", Prompt: "Choose?", Type: template.VariableTypeSelect, Options: []string{"a", "b"}},
			},
		}

		err := v.Validate(tmpl)
		require.NoError(t, err)
	})

	t.Run("empty select option fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "test",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "choice", Prompt: "Choose?", Type: template.VariableTypeSelect, Options: []string{"a", ""}},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be empty")
	})

	t.Run("duplicate select options fail", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "test",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "choice", Prompt: "Choose?", Type: template.VariableTypeSelect, Options: []string{"a", "a"}},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate option")
	})

	t.Run("options on string variable fail", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "test",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName, Options: []string{"a"}},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "options are only allowed")
	})

	t.Run("multiple errors accumulated", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "test",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "var1", Prompt: "", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "var2", Prompt: "Pick?", Type: template.VariableTypeSelect},  // missing options
				{Name: "var2", Prompt: "Again?", Type: template.VariableTypeString}, // duplicate
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		// All three errors should be present
		assert.Contains(t, err.Error(), "Prompt")
		assert.Contains(t, err.Error(), "options required")
		assert.Contains(t, err.Error(), "duplicate variable name")
	})
}

func TestValidator_ValidateProjectNameRole(t *testing.T) {
	v := template.NewValidator()

	t.Run("project template with valid project_name role passes", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "my-project",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "description", Prompt: "Description?", Type: template.VariableTypeString},
			},
		}

		err := v.Validate(tmpl)
		require.NoError(t, err)
	})

	t.Run("project template with zero project_name roles fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "my-project",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString},
				{Name: "description", Prompt: "Description?", Type: template.VariableTypeString},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must have exactly one variable with role")
		assert.Contains(t, err.Error(), "project_name")
	})

	t.Run("project template with no variables fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:      "my-project",
			Type:      template.TypeProject,
			Version:   "1.0.0",
			Variables: []template.Variable{},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must have exactly one variable with role")
		assert.Contains(t, err.Error(), "project_name")
	})

	t.Run("project template with multiple project_name roles fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "my-project",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "project", Prompt: "Project?", Type: template.VariableTypeString, Role: template.RoleProjectName},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has 2 variables with role")
		assert.Contains(t, err.Error(), "must have exactly one")
	})

	t.Run("project template with non-string project_name role fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "my-project",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeBool, Role: template.RoleProjectName},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be of type")
		assert.Contains(t, err.Error(), string(template.VariableTypeString))
	})

	t.Run("feature template without project_name role passes", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "testing-feature",
			Type:    template.TypeFeature,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "use_testify", Prompt: "Use testify?", Type: template.VariableTypeBool},
			},
		}

		err := v.Validate(tmpl)
		require.NoError(t, err)
	})

	t.Run("component template without project_name role passes", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "auth-component",
			Type:    template.TypeComponent,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "provider", Prompt: "Auth provider?", Type: template.VariableTypeString},
			},
		}

		err := v.Validate(tmpl)
		require.NoError(t, err)
	})
}

func TestValidator_ValidateTree(t *testing.T) {
	v := template.NewValidator()

	t.Run("valid tree passes", func(t *testing.T) {
		root := &template.Node{
			Template: &template.Template{
				Name:    "project",
				Type:    template.TypeProject,
				Version: "1.0.0",
				Variables: []template.Variable{
					{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				},
			},
			Children: []*template.Node{
				{
					Template: &template.Template{
						Name:    "feature",
						Type:    template.TypeFeature,
						Version: "1.0.0",
					},
				},
			},
		}

		err := v.ValidateTree(root)
		require.NoError(t, err)
	})

	t.Run("invalid node in tree fails", func(t *testing.T) {
		root := &template.Node{
			Template: &template.Template{
				Name:    "project",
				Type:    template.TypeProject,
				Version: "1.0.0",
				Variables: []template.Variable{
					{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				},
			},
			Children: []*template.Node{
				{
					Template: &template.Template{
						Name:    "", // invalid
						Type:    template.TypeFeature,
						Version: "1.0.0",
					},
				},
			},
		}

		err := v.ValidateTree(root)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Name")
	})

	t.Run("feature including project fails", func(t *testing.T) {
		root := &template.Node{
			Template: &template.Template{
				Name:    "feature",
				Type:    template.TypeFeature,
				Version: "1.0.0",
			},
			Children: []*template.Node{
				{
					Template: &template.Template{
						Name:    "project",
						Type:    template.TypeProject,
						Version: "1.0.0",
						Variables: []template.Variable{
							{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
						},
					},
				},
			},
		}

		err := v.ValidateTree(root)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "feature \"feature\" cannot include project \"project\"")
	})

	t.Run("component including project fails", func(t *testing.T) {
		root := &template.Node{
			Template: &template.Template{
				Name:    "component",
				Type:    template.TypeComponent,
				Version: "1.0.0",
			},
			Children: []*template.Node{
				{
					Template: &template.Template{
						Name:    "project",
						Type:    template.TypeProject,
						Version: "1.0.0",
						Variables: []template.Variable{
							{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
						},
					},
				},
			},
		}

		err := v.ValidateTree(root)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "component \"component\" cannot include project \"project\"")
	})

	t.Run("project including project passes", func(t *testing.T) {
		root := &template.Node{
			Template: &template.Template{
				Name:    "project1",
				Type:    template.TypeProject,
				Version: "1.0.0",
				Variables: []template.Variable{
					{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				},
			},
			Children: []*template.Node{
				{
					Template: &template.Template{
						Name:    "project2",
						Type:    template.TypeProject,
						Version: "1.0.0",
						Variables: []template.Variable{
							{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
						},
					},
				},
			},
		}

		err := v.ValidateTree(root)
		require.NoError(t, err)
	})

	t.Run("duplicate features at same level fail", func(t *testing.T) {
		root := &template.Node{
			Template: &template.Template{
				Name:    "project",
				Type:    template.TypeProject,
				Version: "1.0.0",
				Variables: []template.Variable{
					{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				},
			},
			Children: []*template.Node{
				{
					Template: &template.Template{
						Name:    "testing",
						Type:    template.TypeFeature,
						Version: "1.0.0",
					},
				},
				{
					Template: &template.Template{
						Name:    "testing",
						Type:    template.TypeFeature,
						Version: "1.0.0",
					},
				},
			},
		}

		err := v.ValidateTree(root)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "features and components cannot be included twice at the same level")
		assert.Contains(t, err.Error(), "duplicate feature \"testing\" in \"project\"")
	})

	t.Run("duplicate components at same level fail", func(t *testing.T) {
		root := &template.Node{
			Template: &template.Template{
				Name:    "project",
				Type:    template.TypeProject,
				Version: "1.0.0",
				Variables: []template.Variable{
					{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				},
			},
			Children: []*template.Node{
				{
					Template: &template.Template{
						Name:    "auth",
						Type:    template.TypeComponent,
						Version: "1.0.0",
					},
				},
				{
					Template: &template.Template{
						Name:    "auth",
						Type:    template.TypeComponent,
						Version: "1.0.0",
					},
				},
			},
		}

		err := v.ValidateTree(root)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "features and components cannot be included twice at the same level")
		assert.Contains(t, err.Error(), "duplicate component \"auth\" in \"project\"")
	})

	t.Run("same feature at different levels passes", func(t *testing.T) {
		root := &template.Node{
			Template: &template.Template{
				Name:    "project",
				Type:    template.TypeProject,
				Version: "1.0.0",
				Variables: []template.Variable{
					{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				},
			},
			Children: []*template.Node{
				{
					Template: &template.Template{
						Name:    "testing",
						Type:    template.TypeFeature,
						Version: "1.0.0",
					},
					Children: []*template.Node{
						{
							Template: &template.Template{
								Name:    "testing",
								Type:    template.TypeFeature,
								Version: "1.0.0",
							},
						},
					},
				},
			},
		}

		err := v.ValidateTree(root)
		require.NoError(t, err)
	})
}

func TestValidator_ValidateFiles(t *testing.T) {
	v := template.NewValidator()
	fsys := fstest.MapFS{
		"existing.txt": &fstest.MapFile{Data: []byte("content")},
	}

	t.Run("existing file passes", func(t *testing.T) {
		node := &template.Node{
			Template: &template.Template{
				Name:    "test",
				Type:    template.TypeProject,
				Version: "1.0.0",
				Variables: []template.Variable{
					{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				},
				Files: []template.File{
					{Src: "existing.txt", Dest: "dest.txt"},
				},
			},
			FS:   fsys,
			Path: ".",
		}

		err := v.ValidateTree(node)
		require.NoError(t, err)
	})

	t.Run("missing file fails", func(t *testing.T) {
		node := &template.Node{
			Template: &template.Template{
				Name:    "test",
				Type:    template.TypeProject,
				Version: "1.0.0",
				Variables: []template.Variable{
					{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				},
				Files: []template.File{
					{Src: "missing.txt", Dest: "dest.txt"},
				},
			},
			FS:   fsys,
			Path: ".",
		}

		err := v.ValidateTree(node)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "source file \"missing.txt\" does not exist")
	})
}

func TestValidator_ValidateContext(t *testing.T) {
	v := template.NewValidator()

	tmpl := &template.Template{
		Name: "test",
		Variables: []template.Variable{
			{Name: "required", Prompt: "?", Type: template.VariableTypeString},
			{Name: "optional", Prompt: "?", Type: template.VariableTypeString, Default: "default"},
		},
	}

	t.Run("valid context passes", func(t *testing.T) {
		ctx := template.NewTemplateContext(map[string]any{
			"required": "value",
			"optional": "configured",
		})
		err := v.ValidateContext(tmpl, ctx)
		require.NoError(t, err)
	})

	t.Run("missing required variable fails", func(t *testing.T) {
		ctx := template.NewTemplateContext(map[string]any{})
		err := v.ValidateContext(tmpl, ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "variable required is missing")
	})

	t.Run("missing variable with default still fails", func(t *testing.T) {
		ctx := template.NewTemplateContext(map[string]any{
			"required": "value",
		})
		err := v.ValidateContext(tmpl, ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "variable optional is missing")
	})

	t.Run("invalid string type fails", func(t *testing.T) {
		ctx := template.NewTemplateContext(map[string]any{
			"required": 123,
			"optional": "configured",
		})
		err := v.ValidateContext(tmpl, ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "variable required is invalid")
		assert.Contains(t, err.Error(), "expected type string")
	})

	t.Run("int and bool pass", func(t *testing.T) {
		typed := &template.Template{
			Name: "typed",
			Variables: []template.Variable{
				{Name: "port", Prompt: "?", Type: template.VariableTypeInt},
				{Name: "enabled", Prompt: "?", Type: template.VariableTypeBool},
			},
		}
		ctx := template.NewTemplateContext(map[string]any{
			"port":    8080,
			"enabled": true,
		})
		err := v.ValidateContext(typed, ctx)
		require.NoError(t, err)
	})

	t.Run("string int fails", func(t *testing.T) {
		typed := &template.Template{
			Name: "typed",
			Variables: []template.Variable{
				{Name: "port", Prompt: "?", Type: template.VariableTypeInt},
			},
		}
		ctx := template.NewTemplateContext(map[string]any{
			"port": "8080",
		})
		err := v.ValidateContext(typed, ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected type int")
	})

	t.Run("non-int value fails", func(t *testing.T) {
		typed := &template.Template{
			Name: "typed",
			Variables: []template.Variable{
				{Name: "port", Prompt: "?", Type: template.VariableTypeInt},
			},
		}
		ctx := template.NewTemplateContext(map[string]any{
			"port": 3.14,
		})
		err := v.ValidateContext(typed, ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected type int")
	})

	t.Run("non-bool bool fails", func(t *testing.T) {
		typed := &template.Template{
			Name: "typed",
			Variables: []template.Variable{
				{Name: "enabled", Prompt: "?", Type: template.VariableTypeBool},
			},
		}
		ctx := template.NewTemplateContext(map[string]any{
			"enabled": "true",
		})
		err := v.ValidateContext(typed, ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected type bool")
	})

	t.Run("select validates string values", func(t *testing.T) {
		typed := &template.Template{
			Name: "typed",
			Variables: []template.Variable{
				{Name: "color", Prompt: "?", Type: template.VariableTypeSelect, Options: []string{"red", "blue"}},
			},
		}
		ctx := template.NewTemplateContext(map[string]any{
			"color": "red",
		})
		err := v.ValidateContext(typed, ctx)
		require.NoError(t, err)
	})

	t.Run("select value outside options fails", func(t *testing.T) {
		typed := &template.Template{
			Name: "typed",
			Variables: []template.Variable{
				{Name: "color", Prompt: "?", Type: template.VariableTypeSelect, Options: []string{"red", "blue"}},
			},
		}
		ctx := template.NewTemplateContext(map[string]any{
			"color": "green",
		})
		err := v.ValidateContext(typed, ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "contains invalid option \"green\"")
	})

	t.Run("select requires string values", func(t *testing.T) {
		typed := &template.Template{
			Name: "typed",
			Variables: []template.Variable{
				{Name: "color", Prompt: "?", Type: template.VariableTypeSelect, Options: []string{"red", "blue"}},
			},
		}
		ctx := template.NewTemplateContext(map[string]any{
			"color": []string{"red"},
		})
		err := v.ValidateContext(typed, ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected type select")
	})

	t.Run("multiselect validates slice values", func(t *testing.T) {
		typed := &template.Template{
			Name: "typed",
			Variables: []template.Variable{
				{Name: "features", Prompt: "?", Type: template.VariableTypeMultiSelect, Options: []string{"api", "db"}},
			},
		}
		ctx := template.NewTemplateContext(map[string]any{
			"features": []string{"api", "db"},
		})
		err := v.ValidateContext(typed, ctx)
		require.NoError(t, err)
	})

	t.Run("multiselect with invalid option fails", func(t *testing.T) {
		typed := &template.Template{
			Name: "typed",
			Variables: []template.Variable{
				{Name: "features", Prompt: "?", Type: template.VariableTypeMultiSelect, Options: []string{"api", "db"}},
			},
		}
		ctx := template.NewTemplateContext(map[string]any{
			"features": []any{"api", "cache"},
		})
		err := v.ValidateContext(typed, ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "contains invalid option \"cache\"")
	})
}

func TestValidator_Validate_DefaultTypes(t *testing.T) {
	v := template.NewValidator()

	t.Run("int default and bool default pass", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "typed",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "port", Prompt: "?", Type: template.VariableTypeInt, Default: 8080},
				{Name: "enabled", Prompt: "?", Type: template.VariableTypeBool, Default: true},
			},
		}

		err := v.Validate(tmpl)
		require.NoError(t, err)
	})

	t.Run("invalid default type fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "typed",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "port", Prompt: "?", Type: template.VariableTypeInt, Default: "8080"},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid default value")
		assert.Contains(t, err.Error(), "expected type int")
	})

	t.Run("string bool default fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "typed",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "enabled", Prompt: "?", Type: template.VariableTypeBool, Default: "true"},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid default value")
		assert.Contains(t, err.Error(), "expected type bool")
	})

	t.Run("select default must be string", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "typed",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "color", Prompt: "?", Type: template.VariableTypeSelect, Options: []string{"red", "blue"}, Default: "red"},
			},
		}

		err := v.Validate(tmpl)
		require.NoError(t, err)
	})

	t.Run("invalid multiselect default option fails", func(t *testing.T) {
		tmpl := &template.Template{
			Name:    "typed",
			Type:    template.TypeProject,
			Version: "1.0.0",
			Variables: []template.Variable{
				{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "features", Prompt: "?", Type: template.VariableTypeMultiSelect, Options: []string{"api", "db"}, Default: []any{"api", "cache"}},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid default value")
		assert.Contains(t, err.Error(), "contains invalid option")
	})
}

func TestValidator_ValidateTreeContexts(t *testing.T) {
	v := template.NewValidator()

	root := &template.Node{
		ID: "0",
		Template: &template.Template{
			Name: "root",
			Variables: []template.Variable{
				{Name: "var_root", Prompt: "?", Type: template.VariableTypeString},
			},
		},
		Children: []*template.Node{
			{
				ID: "0.0",
				Template: &template.Template{
					Name: "child",
					Variables: []template.Variable{
						{Name: "var_child", Prompt: "?", Type: template.VariableTypeString},
					},
				},
			},
		},
	}

	t.Run("valid contexts pass", func(t *testing.T) {
		contexts := template.RenderContexts{
			"0":   template.NewTemplateContext(map[string]any{"var_root": "val"}),
			"0.0": template.NewTemplateContext(map[string]any{"var_child": "val"}),
		}
		err := v.ValidateTreeContexts(root, contexts)
		require.NoError(t, err)
	})

	t.Run("missing context for a node fails", func(t *testing.T) {
		contexts := template.RenderContexts{
			"0": template.NewTemplateContext(map[string]any{"var_root": "val"}),
		}
		err := v.ValidateTreeContexts(root, contexts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no context found for template child (ID: 0.0)")
	})

	t.Run("missing variable in one of the contexts fails", func(t *testing.T) {
		contexts := template.RenderContexts{
			"0":   template.NewTemplateContext(map[string]any{"var_root": "val"}),
			"0.0": template.NewTemplateContext(map[string]any{}), // missing var_child
		}
		err := v.ValidateTreeContexts(root, contexts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "variable var_child is missing")
	})
}
