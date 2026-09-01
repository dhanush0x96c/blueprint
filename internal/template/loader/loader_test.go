// Package loader_test contains unit tests for the loader package.
package loader_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/template/loader"
	"github.com/dhanush0x96c/blueprint/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestLoader_Load(t *testing.T) {
	base := t.TempDir()
	fsys := os.DirFS(base)
	l := loader.NewLoader()

	t.Run("load from relative directory", func(t *testing.T) {
		dir := filepath.Join(base, "projects", "go-cli")
		testutil.WriteTemplateStruct(t, dir, testutil.NewTemplate("go-cli",
			testutil.WithVariable(template.Variable{Name: "app", Role: template.RoleProjectName, Prompt: "?", Type: template.VariableTypeString})))

		tmpl, err := l.Load(fsys, "projects/go-cli")
		require.NoError(t, err)
		require.Equal(t, "go-cli", tmpl.Template.Name)
	})

	t.Run("load from template.yaml path", func(t *testing.T) {
		dir := filepath.Join(base, "direct")
		testutil.WriteTemplateStruct(t, dir, testutil.NewTemplate("go-cli",
			testutil.WithVariable(template.Variable{Name: "app", Role: template.RoleProjectName, Prompt: "?", Type: template.VariableTypeString})))

		path := filepath.Join("direct", loader.FileName)
		tmpl, err := l.Load(fsys, path)
		require.NoError(t, err)
		require.Equal(t, "go-cli", tmpl.Template.Name)
	})

	t.Run("invalid template fails validation", func(t *testing.T) {
		templateName := "invalid"
		dir := filepath.Join(base, templateName)
		// Version is required but missing here
		testutil.WriteTemplateStruct(t, dir, &template.Template{Name: "invalid", Type: template.TypeProject})

		_, err := l.Load(fsys, templateName)
		require.Error(t, err)
	})
}

func TestLoader_LoadTags(t *testing.T) {
	base := t.TempDir()
	fsys := os.DirFS(base)
	l := loader.NewLoader()

	t.Run("loads tags when present", func(t *testing.T) {
		dir := filepath.Join(base, "with-tags")
		testutil.WriteTemplateStruct(t, dir, testutil.NewTemplate("tagged-template",
			testutil.WithVariable(template.Variable{Name: "app", Role: template.RoleProjectName, Prompt: "?", Type: template.VariableTypeString}),
			testutil.WithTags("go", "cli", "testing")))

		tmpl, err := l.Load(fsys, "with-tags")
		require.NoError(t, err)
		require.Equal(t, "tagged-template", tmpl.Template.Name)
		require.Len(t, tmpl.Template.Tags, 3)
		require.Equal(t, []string{"go", "cli", "testing"}, tmpl.Template.Tags)
	})

	t.Run("handles missing tags", func(t *testing.T) {
		dir := filepath.Join(base, "without-tags")
		testutil.WriteTemplateStruct(t, dir, testutil.NewTemplate("no-tags",
			testutil.WithType(template.TypeFeature)))

		tmpl, err := l.Load(fsys, "without-tags")
		require.NoError(t, err)
		require.Equal(t, "no-tags", tmpl.Template.Name)
		require.Nil(t, tmpl.Template.Tags)
	})
}
