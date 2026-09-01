// Package validator_test contains unit tests for the validator package.
package validator_test

import (
	"testing"
	"testing/fstest"

	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/template/validator"
	"github.com/dhanush0x96c/blueprint/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_Validate(t *testing.T) {
	v := validator.NewValidator()

	testCases := []struct {
		name        string
		tmpl        *template.Template
		wantErr     bool
		errContains []string
	}{
		{
			name: "valid template passes",
			tmpl: testutil.NewTemplate("test",
				testutil.WithVariable(template.Variable{
					Name:   "app_name",
					Prompt: "App name?",
					Type:   template.VariableTypeString,
					Role:   template.RoleProjectName,
				})),
			wantErr: false,
		},
		{
			name: "missing required fields fails",
			tmpl: testutil.NewTemplate("",
				testutil.WithType(template.TypeProject),
				testutil.WithVersion("")),
			wantErr:     true,
			errContains: []string{"Name", "Version"},
		},
		{
			name: "invalid type fails",
			tmpl: testutil.NewTemplate("test",
				testutil.WithType("invalid")),
			wantErr:     true,
			errContains: []string{"Type"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate(tc.tmpl)
			if tc.wantErr {
				require.Error(t, err)
				for _, msg := range tc.errContains {
					assert.Contains(t, err.Error(), msg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidator_ValidateVariables(t *testing.T) {
	v := validator.NewValidator()

	testCases := []struct {
		name        string
		variables   []template.Variable
		wantErr     bool
		errContains []string
	}{
		{
			name: "duplicate variable names fails",
			variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "app_name", Prompt: "Another?", Type: template.VariableTypeString},
			},
			wantErr:     true,
			errContains: []string{"duplicate variable name", "app_name"},
		},
		{
			name: "missing prompt fails",
			variables: []template.Variable{
				{Name: "app_name", Prompt: "", Type: template.VariableTypeString},
			},
			wantErr:     true,
			errContains: []string{"Prompt"},
		},
		{
			name: "select without options fails",
			variables: []template.Variable{
				{Name: "choice", Prompt: "Choose?", Type: template.VariableTypeSelect, Options: []string{}},
			},
			wantErr:     true,
			errContains: []string{"options required"},
		},
		{
			name: "multiselect without options fails",
			variables: []template.Variable{
				{Name: "choices", Prompt: "Choose?", Type: template.VariableTypeMultiSelect, Options: nil},
			},
			wantErr:     true,
			errContains: []string{"options required"},
		},
		{
			name: "select with options passes",
			variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "choice", Prompt: "Choose?", Type: template.VariableTypeSelect, Options: []string{"a", "b"}},
			},
			wantErr: false,
		},
		{
			name: "empty select option fails",
			variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "choice", Prompt: "Choose?", Type: template.VariableTypeSelect, Options: []string{"a", ""}},
			},
			wantErr:     true,
			errContains: []string{"must not be empty"},
		},
		{
			name: "duplicate select options fail",
			variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "choice", Prompt: "Choose?", Type: template.VariableTypeSelect, Options: []string{"a", "a"}},
			},
			wantErr:     true,
			errContains: []string{"duplicate option"},
		},
		{
			name: "options on string variable fail",
			variables: []template.Variable{
				{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName, Options: []string{"a"}},
			},
			wantErr:     true,
			errContains: []string{"options are only allowed"},
		},
		{
			name: "multiple errors accumulated",
			variables: []template.Variable{
				{Name: "var1", Prompt: "", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "var2", Prompt: "Pick?", Type: template.VariableTypeSelect},  // missing options
				{Name: "var2", Prompt: "Again?", Type: template.VariableTypeString}, // duplicate
			},
			wantErr:     true,
			errContains: []string{"Prompt", "options required", "duplicate variable name"},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			opts := []testutil.TemplateOption{testutil.WithType(template.TypeProject)}
			for _, v := range tc.variables {
				opts = append(opts, testutil.WithVariable(v))
			}
			tmpl := testutil.NewTemplate("test", opts...)

			err := v.Validate(tmpl)
			if tc.wantErr {
				require.Error(t, err)
				for _, msg := range tc.errContains {
					assert.Contains(t, err.Error(), msg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidator_ValidateProjectNameRole(t *testing.T) {
	v := validator.NewValidator()

	testCases := []struct {
		name        string
		tmpl        *template.Template
		wantErr     bool
		errContains []string
	}{
		{
			name: "project template with valid project_name role passes",
			tmpl: testutil.NewTemplate("my-project",
				testutil.WithVariable(template.Variable{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName}),
				testutil.WithVariable(template.Variable{Name: "description", Prompt: "Description?", Type: template.VariableTypeString})),
			wantErr: false,
		},
		{
			name: "project template with zero project_name roles fails",
			tmpl: testutil.NewTemplate("my-project",
				testutil.WithVariable(template.Variable{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString}),
				testutil.WithVariable(template.Variable{Name: "description", Prompt: "Description?", Type: template.VariableTypeString})),
			wantErr:     true,
			errContains: []string{"must have exactly one variable with role", "project_name"},
		},
		{
			name:        "project template with no variables fails",
			tmpl:        testutil.NewTemplate("my-project"),
			wantErr:     true,
			errContains: []string{"must have exactly one variable with role", "project_name"},
		},
		{
			name: "project template with multiple project_name roles fails",
			tmpl: testutil.NewTemplate("my-project",
				testutil.WithVariable(template.Variable{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeString, Role: template.RoleProjectName}),
				testutil.WithVariable(template.Variable{Name: "project", Prompt: "Project?", Type: template.VariableTypeString, Role: template.RoleProjectName})),
			wantErr:     true,
			errContains: []string{"has 2 variables with role", "must have exactly one"},
		},
		{
			name: "project template with non-string project_name role fails",
			tmpl: testutil.NewTemplate("my-project",
				testutil.WithVariable(template.Variable{Name: "app_name", Prompt: "App name?", Type: template.VariableTypeBool, Role: template.RoleProjectName})),
			wantErr:     true,
			errContains: []string{"must be of type", string(template.VariableTypeString)},
		},
		{
			name: "feature template without project_name role passes",
			tmpl: testutil.NewTemplate("testing-feature",
				testutil.WithType(template.TypeFeature),
				testutil.WithVariable(template.Variable{Name: "use_testify", Prompt: "Use testify?", Type: template.VariableTypeBool})),
			wantErr: false,
		},
		{
			name: "component template without project_name role passes",
			tmpl: testutil.NewTemplate("auth-component",
				testutil.WithType(template.TypeComponent),
				testutil.WithVariable(template.Variable{Name: "provider", Prompt: "Auth provider?", Type: template.VariableTypeString})),
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate(tc.tmpl)
			if tc.wantErr {
				require.Error(t, err)
				for _, msg := range tc.errContains {
					assert.Contains(t, err.Error(), msg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidator_ValidateTree(t *testing.T) {
	v := validator.NewValidator()
	projectWithNameVar := func(name string) *template.Template {
		return testutil.NewTemplate(name,
			testutil.WithVariable(template.Variable{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName}))
	}
	feature := func(name string) *template.Template {
		return testutil.NewTemplate(name, testutil.WithType(template.TypeFeature))
	}
	component := func(name string) *template.Template {
		return testutil.NewTemplate(name, testutil.WithType(template.TypeComponent))
	}

	testCases := []struct {
		name        string
		root        *template.Node
		wantErr     bool
		errContains []string
	}{
		{
			name: "valid tree passes",
			root: testutil.NewNode("0", projectWithNameVar("project"),
				testutil.WithChild(testutil.NewNode("0.0", feature("feature")))),
			wantErr: false,
		},
		{
			name: "invalid node in tree fails",
			root: testutil.NewNode("0", projectWithNameVar("project"),
				testutil.WithChild(testutil.NewNode("0.0", feature("")))),
			wantErr:     true,
			errContains: []string{"Name"},
		},
		{
			name: "feature including project fails",
			root: testutil.NewNode("0", feature("feature"),
				testutil.WithChild(testutil.NewNode("0.0", projectWithNameVar("project")))),
			wantErr:     true,
			errContains: []string{"feature \"feature\" cannot include project \"project\""},
		},
		{
			name: "component including project fails",
			root: testutil.NewNode("0", component("component"),
				testutil.WithChild(testutil.NewNode("0.0", projectWithNameVar("project")))),
			wantErr:     true,
			errContains: []string{"component \"component\" cannot include project \"project\""},
		},
		{
			name: "project including project passes",
			root: testutil.NewNode("0", projectWithNameVar("project1"),
				testutil.WithChild(testutil.NewNode("0.0", projectWithNameVar("project2")))),
			wantErr: false,
		},
		{
			name: "duplicate features at same level fail",
			root: testutil.NewNode("0", projectWithNameVar("project"),
				testutil.WithChild(testutil.NewNode("0.0", feature("testing"))),
				testutil.WithChild(testutil.NewNode("0.1", feature("testing")))),
			wantErr:     true,
			errContains: []string{"features and components cannot be included twice at the same level", "duplicate feature \"testing\" in \"project\""},
		},
		{
			name: "duplicate components at same level fail",
			root: testutil.NewNode("0", projectWithNameVar("project"),
				testutil.WithChild(testutil.NewNode("0.0", component("auth"))),
				testutil.WithChild(testutil.NewNode("0.1", component("auth")))),
			wantErr:     true,
			errContains: []string{"features and components cannot be included twice at the same level", "duplicate component \"auth\" in \"project\""},
		},
		{
			name: "same feature at different levels passes",
			root: testutil.NewNode("0", projectWithNameVar("project"),
				testutil.WithChild(
					testutil.NewNode("0.0", feature("testing"),
						testutil.WithChild(testutil.NewNode("0.0.0", feature("testing")))))),
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateTree(tc.root)
			if tc.wantErr {
				require.Error(t, err)
				for _, msg := range tc.errContains {
					assert.Contains(t, err.Error(), msg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidator_ValidateFiles(t *testing.T) {
	v := validator.NewValidator()
	fsys := fstest.MapFS{
		"existing.txt": &fstest.MapFile{Data: []byte("content")},
	}

	t.Run("existing file passes", func(t *testing.T) {
		tmpl := testutil.NewTemplate("test",
			testutil.WithVariable(template.Variable{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName}),
			testutil.WithFile(template.File{Src: "existing.txt", Dest: "dest.txt"}))

		node := testutil.NewNode("0", tmpl,
			testutil.WithFS(fsys),
			testutil.WithPath("."))

		err := v.ValidateTree(node)
		require.NoError(t, err)
	})

	t.Run("missing file fails", func(t *testing.T) {
		tmpl := testutil.NewTemplate("test",
			testutil.WithVariable(template.Variable{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName}),
			testutil.WithFile(template.File{Src: "missing.txt", Dest: "dest.txt"}))

		node := testutil.NewNode("0", tmpl,
			testutil.WithFS(fsys),
			testutil.WithPath("."))

		err := v.ValidateTree(node)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "source file \"missing.txt\" does not exist")
	})
}

func TestValidator_ValidateContext(t *testing.T) {
	v := validator.NewValidator()

	baseTemplate := testutil.NewTemplate("test",
		testutil.WithVariable(template.Variable{Name: "required", Prompt: "?", Type: template.VariableTypeString}),
		testutil.WithVariable(template.Variable{Name: "optional", Prompt: "?", Type: template.VariableTypeString, Default: "default"}))

	typedTemplate := func(vars []template.Variable) *template.Template {
		opts := []testutil.TemplateOption{}
		for _, v := range vars {
			opts = append(opts, testutil.WithVariable(v))
		}
		return testutil.NewTemplate("typed", opts...)
	}

	testCases := []struct {
		name        string
		tmpl        *template.Template
		values      map[string]any
		wantErr     bool
		errContains []string
	}{
		{
			name: "valid context passes",
			tmpl: baseTemplate,
			values: map[string]any{
				"required": "value",
				"optional": "configured",
			},
			wantErr: false,
		},
		{
			name:        "missing required variable fails",
			tmpl:        baseTemplate,
			values:      map[string]any{},
			wantErr:     true,
			errContains: []string{"variable required is missing"},
		},
		{
			name: "missing variable with default still fails",
			tmpl: baseTemplate,
			values: map[string]any{
				"required": "value",
			},
			wantErr:     true,
			errContains: []string{"variable optional is missing"},
		},
		{
			name: "invalid string type fails",
			tmpl: baseTemplate,
			values: map[string]any{
				"required": 123,
				"optional": "configured",
			},
			wantErr:     true,
			errContains: []string{"variable required is invalid", "expected type string"},
		},
		{
			name: "int and bool pass",
			tmpl: typedTemplate([]template.Variable{
				{Name: "port", Prompt: "?", Type: template.VariableTypeInt},
				{Name: "enabled", Prompt: "?", Type: template.VariableTypeBool},
			}),
			values: map[string]any{
				"port":    8080,
				"enabled": true,
			},
			wantErr: false,
		},
		{
			name: "string int fails",
			tmpl: typedTemplate([]template.Variable{
				{Name: "port", Prompt: "?", Type: template.VariableTypeInt},
			}),
			values: map[string]any{
				"port": "8080",
			},
			wantErr:     true,
			errContains: []string{"expected type int"},
		},
		{
			name: "non-int value fails",
			tmpl: typedTemplate([]template.Variable{
				{Name: "port", Prompt: "?", Type: template.VariableTypeInt},
			}),
			values: map[string]any{
				"port": 3.14,
			},
			wantErr:     true,
			errContains: []string{"expected type int"},
		},
		{
			name: "non-bool bool fails",
			tmpl: typedTemplate([]template.Variable{
				{Name: "enabled", Prompt: "?", Type: template.VariableTypeBool},
			}),
			values: map[string]any{
				"enabled": "true",
			},
			wantErr:     true,
			errContains: []string{"expected type bool"},
		},
		{
			name: "select validates string values",
			tmpl: typedTemplate([]template.Variable{
				{Name: "color", Prompt: "?", Type: template.VariableTypeSelect, Options: []string{"red", "blue"}},
			}),
			values: map[string]any{
				"color": "red",
			},
			wantErr: false,
		},
		{
			name: "select value outside options fails",
			tmpl: typedTemplate([]template.Variable{
				{Name: "color", Prompt: "?", Type: template.VariableTypeSelect, Options: []string{"red", "blue"}},
			}),
			values: map[string]any{
				"color": "green",
			},
			wantErr:     true,
			errContains: []string{"contains invalid option \"green\""},
		},
		{
			name: "select requires string values",
			tmpl: typedTemplate([]template.Variable{
				{Name: "color", Prompt: "?", Type: template.VariableTypeSelect, Options: []string{"red", "blue"}},
			}),
			values: map[string]any{
				"color": []string{"red"},
			},
			wantErr:     true,
			errContains: []string{"expected type select"},
		},
		{
			name: "multiselect validates slice values",
			tmpl: typedTemplate([]template.Variable{
				{Name: "features", Prompt: "?", Type: template.VariableTypeMultiSelect, Options: []string{"api", "db"}},
			}),
			values: map[string]any{
				"features": []string{"api", "db"},
			},
			wantErr: false,
		},
		{
			name: "multiselect with invalid option fails",
			tmpl: typedTemplate([]template.Variable{
				{Name: "features", Prompt: "?", Type: template.VariableTypeMultiSelect, Options: []string{"api", "db"}},
			}),
			values: map[string]any{
				"features": []any{"api", "cache"},
			},
			wantErr:     true,
			errContains: []string{"contains invalid option \"cache\""},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := template.NewTemplateContext(tc.values)
			err := v.ValidateContext(tc.tmpl, ctx)
			if tc.wantErr {
				require.Error(t, err)
				for _, msg := range tc.errContains {
					assert.Contains(t, err.Error(), msg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidator_Validate_DefaultTypes(t *testing.T) {
	v := validator.NewValidator()

	testCases := []struct {
		name        string
		variables   []template.Variable
		wantErr     bool
		errContains []string
	}{
		{
			name: "int default and bool default pass",
			variables: []template.Variable{
				{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "port", Prompt: "?", Type: template.VariableTypeInt, Default: 8080},
				{Name: "enabled", Prompt: "?", Type: template.VariableTypeBool, Default: true},
			},
			wantErr: false,
		},
		{
			name: "invalid default type fails",
			variables: []template.Variable{
				{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "port", Prompt: "?", Type: template.VariableTypeInt, Default: "8080"},
			},
			wantErr:     true,
			errContains: []string{"invalid default value", "expected type int"},
		},
		{
			name: "string bool default fails",
			variables: []template.Variable{
				{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "enabled", Prompt: "?", Type: template.VariableTypeBool, Default: "true"},
			},
			wantErr:     true,
			errContains: []string{"invalid default value", "expected type bool"},
		},
		{
			name: "select default must be string",
			variables: []template.Variable{
				{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "color", Prompt: "?", Type: template.VariableTypeSelect, Options: []string{"red", "blue"}, Default: "red"},
			},
			wantErr: false,
		},
		{
			name: "invalid multiselect default option fails",
			variables: []template.Variable{
				{Name: "app", Prompt: "?", Type: template.VariableTypeString, Role: template.RoleProjectName},
				{Name: "features", Prompt: "?", Type: template.VariableTypeMultiSelect, Options: []string{"api", "db"}, Default: []any{"api", "cache"}},
			},
			wantErr:     true,
			errContains: []string{"invalid default value", "contains invalid option"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tmpl := &template.Template{
				Name:      "typed",
				Type:      template.TypeProject,
				Version:   "1.0.0",
				Variables: tc.variables,
			}

			err := v.Validate(tmpl)
			if tc.wantErr {
				require.Error(t, err)
				for _, msg := range tc.errContains {
					assert.Contains(t, err.Error(), msg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidator_ValidateTreeContexts(t *testing.T) {
	v := validator.NewValidator()

	root := testutil.NewNode("0",
		testutil.NewTemplate("root",
			testutil.WithVariable(template.Variable{Name: "var_root", Prompt: "?", Type: template.VariableTypeString})),
		testutil.WithChild(
			testutil.NewNode("0.0",
				testutil.NewTemplate("child",
					testutil.WithVariable(template.Variable{Name: "var_child", Prompt: "?", Type: template.VariableTypeString})))))

	testCases := []struct {
		name        string
		contexts    template.RenderContexts
		wantErr     bool
		errContains []string
	}{
		{
			name: "valid contexts pass",
			contexts: template.RenderContexts{
				"0":   template.NewTemplateContext(map[string]any{"var_root": "val"}),
				"0.0": template.NewTemplateContext(map[string]any{"var_child": "val"}),
			},
			wantErr: false,
		},
		{
			name: "missing context for a node fails",
			contexts: template.RenderContexts{
				"0": template.NewTemplateContext(map[string]any{"var_root": "val"}),
			},
			wantErr:     true,
			errContains: []string{"no context found for template child (ID: 0.0)"},
		},
		{
			name: "missing variable in one of the contexts fails",
			contexts: template.RenderContexts{
				"0":   template.NewTemplateContext(map[string]any{"var_root": "val"}),
				"0.0": template.NewTemplateContext(map[string]any{}),
			},
			wantErr:     true,
			errContains: []string{"variable var_child is missing"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateTreeContexts(root, tc.contexts)
			if tc.wantErr {
				require.Error(t, err)
				for _, msg := range tc.errContains {
					assert.Contains(t, err.Error(), msg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}
