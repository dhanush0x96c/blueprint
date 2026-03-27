package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_Validate(t *testing.T) {
	v := NewValidator()

	t.Run("valid template passes", func(t *testing.T) {
		tmpl := &Template{
			Name:    "test",
			Type:    TypeProject,
			Version: "1.0.0",
			Variables: []Variable{
				{Name: "app_name", Prompt: "App name?", Type: VariableTypeString, Role: RoleProjectName},
			},
		}

		err := v.Validate(tmpl)
		require.NoError(t, err)
	})

	t.Run("missing required fields fails", func(t *testing.T) {
		tmpl := &Template{
			Name: "", // missing
			Type: TypeProject,
			// Version missing
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Name")
		assert.Contains(t, err.Error(), "Version")
	})

	t.Run("invalid type fails", func(t *testing.T) {
		tmpl := &Template{
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
	v := NewValidator()

	t.Run("duplicate variable names fails", func(t *testing.T) {
		tmpl := &Template{
			Name:    "test",
			Type:    TypeProject,
			Version: "1.0.0",
			Variables: []Variable{
				{Name: "app_name", Prompt: "App name?", Type: VariableTypeString, Role: RoleProjectName},
				{Name: "app_name", Prompt: "Another?", Type: VariableTypeString},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate variable name")
		assert.Contains(t, err.Error(), "app_name")
	})

	t.Run("missing prompt fails", func(t *testing.T) {
		tmpl := &Template{
			Name:    "test",
			Type:    TypeProject,
			Version: "1.0.0",
			Variables: []Variable{
				{Name: "app_name", Prompt: "", Type: VariableTypeString},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Prompt")
	})

	t.Run("select without options fails", func(t *testing.T) {
		tmpl := &Template{
			Name:    "test",
			Type:    TypeProject,
			Version: "1.0.0",
			Variables: []Variable{
				{Name: "choice", Prompt: "Choose?", Type: VariableTypeSelect, Options: []string{}},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "options required")
	})

	t.Run("multiselect without options fails", func(t *testing.T) {
		tmpl := &Template{
			Name:    "test",
			Type:    TypeProject,
			Version: "1.0.0",
			Variables: []Variable{
				{Name: "choices", Prompt: "Choose?", Type: VariableTypeMultiSelect, Options: nil},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "options required")
	})

	t.Run("select with options passes", func(t *testing.T) {
		tmpl := &Template{
			Name:    "test",
			Type:    TypeProject,
			Version: "1.0.0",
			Variables: []Variable{
				{Name: "app_name", Prompt: "App name?", Type: VariableTypeString, Role: RoleProjectName},
				{Name: "choice", Prompt: "Choose?", Type: VariableTypeSelect, Options: []string{"a", "b"}},
			},
		}

		err := v.Validate(tmpl)
		require.NoError(t, err)
	})

	t.Run("multiple errors accumulated", func(t *testing.T) {
		tmpl := &Template{
			Name:    "test",
			Type:    TypeProject,
			Version: "1.0.0",
			Variables: []Variable{
				{Name: "var1", Prompt: "", Type: VariableTypeString, Role: RoleProjectName},
				{Name: "var2", Prompt: "Pick?", Type: VariableTypeSelect},  // missing options
				{Name: "var2", Prompt: "Again?", Type: VariableTypeString}, // duplicate
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
	v := NewValidator()

	t.Run("project template with valid project_name role passes", func(t *testing.T) {
		tmpl := &Template{
			Name:    "my-project",
			Type:    TypeProject,
			Version: "1.0.0",
			Variables: []Variable{
				{Name: "app_name", Prompt: "App name?", Type: VariableTypeString, Role: RoleProjectName},
				{Name: "description", Prompt: "Description?", Type: VariableTypeString},
			},
		}

		err := v.Validate(tmpl)
		require.NoError(t, err)
	})

	t.Run("project template with zero project_name roles fails", func(t *testing.T) {
		tmpl := &Template{
			Name:    "my-project",
			Type:    TypeProject,
			Version: "1.0.0",
			Variables: []Variable{
				{Name: "app_name", Prompt: "App name?", Type: VariableTypeString},
				{Name: "description", Prompt: "Description?", Type: VariableTypeString},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must have exactly one variable with role")
		assert.Contains(t, err.Error(), "project_name")
	})

	t.Run("project template with no variables fails", func(t *testing.T) {
		tmpl := &Template{
			Name:      "my-project",
			Type:      TypeProject,
			Version:   "1.0.0",
			Variables: []Variable{},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must have exactly one variable with role")
		assert.Contains(t, err.Error(), "project_name")
	})

	t.Run("project template with multiple project_name roles fails", func(t *testing.T) {
		tmpl := &Template{
			Name:    "my-project",
			Type:    TypeProject,
			Version: "1.0.0",
			Variables: []Variable{
				{Name: "app_name", Prompt: "App name?", Type: VariableTypeString, Role: RoleProjectName},
				{Name: "project", Prompt: "Project?", Type: VariableTypeString, Role: RoleProjectName},
			},
		}

		err := v.Validate(tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has 2 variables with role")
		assert.Contains(t, err.Error(), "must have exactly one")
	})

	t.Run("feature template without project_name role passes", func(t *testing.T) {
		tmpl := &Template{
			Name:    "testing-feature",
			Type:    TypeFeature,
			Version: "1.0.0",
			Variables: []Variable{
				{Name: "use_testify", Prompt: "Use testify?", Type: VariableTypeBool},
			},
		}

		err := v.Validate(tmpl)
		require.NoError(t, err)
	})

	t.Run("component template without project_name role passes", func(t *testing.T) {
		tmpl := &Template{
			Name:    "auth-component",
			Type:    TypeComponent,
			Version: "1.0.0",
			Variables: []Variable{
				{Name: "provider", Prompt: "Auth provider?", Type: VariableTypeString},
			},
		}

		err := v.Validate(tmpl)
		require.NoError(t, err)
	})
}
