package template_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderer_RenderString(t *testing.T) {
	r, _ := testutil.NewRenderer(t)

	t.Run("simple interpolates variables", func(t *testing.T) {
		out, err := r.RenderString(
			"Hello {{ .name }}",
			testutil.Context(map[string]any{
				"name": "Blueprint",
			}),
			"test",
		)

		require.NoError(t, err)
		assert.Equal(t, "Hello Blueprint", string(out))
	})

	t.Run("applies default template funcs", func(t *testing.T) {
		out, err := r.RenderString(
			"{{ .name | toUpper }}",
			testutil.Context(map[string]any{
				"name": "blueprint",
			}),
			"test",
		)

		require.NoError(t, err)
		assert.Equal(t, "BLUEPRINT", string(out))
	})

	t.Run("parse error", func(t *testing.T) {
		_, err := r.RenderString(
			"{{ .name ",
			testutil.Context(map[string]any{
				"name": "oops",
			}),
			"broken",
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse template")
	})

	t.Run("execution error", func(t *testing.T) {
		_, err := r.RenderString(
			"{{ toInt .value }}",
			testutil.Context(map[string]any{
				"value": "not-a-number",
			}),
			"exec-error",
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute template")
	})
}

func TestRenderer_Render(t *testing.T) {
	r, dir := testutil.NewRenderer(t)

	t.Run("renders .tmpl file", func(t *testing.T) {
		path := filepath.Join(dir, "hello.tmpl")
		err := os.WriteFile(path, []byte("Hi {{ .name }}"), 0644)
		require.NoError(t, err)

		out, err := r.Render(
			os.DirFS(dir),
			"hello.tmpl",
			testutil.Context(map[string]any{
				"name": "World",
			}),
		)

		require.NoError(t, err)
		assert.Equal(t, "Hi World", string(out))
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := r.Render(
			os.DirFS(dir),
			"missing.tmpl",
			testutil.Context(map[string]any{}),
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read template file")
	})
}

func TestRenderer_RenderPath(t *testing.T) {
	r, _ := testutil.NewRenderer(t)

	t.Run("renders path with variables", func(t *testing.T) {
		out, err := r.RenderPath(
			"{{ .pkg }}/main.go",
			testutil.Context(map[string]any{
				"pkg": "myapp",
			}),
		)

		require.NoError(t, err)
		assert.Equal(t, "myapp/main.go", out)
	})
}

func TestRenderer_AddFunc(t *testing.T) {
	r, _ := testutil.NewRenderer(t)

	t.Run("custom function", func(t *testing.T) {
		r.AddFunc("shout", func(s string) string {
			return s + "!!!"
		})

		out, err := r.RenderString(
			"{{ shout .msg }}",
			testutil.Context(map[string]any{
				"msg": "hey",
			}),
			"custom-func",
		)

		require.NoError(t, err)
		assert.Equal(t, "hey!!!", string(out))
	})
}

func TestRenderer_RenderAll(t *testing.T) {
	r, dir := testutil.NewRenderer(t)

	t.Run("renders all files with contexts", func(t *testing.T) {
		err := os.WriteFile(
			filepath.Join(dir, "a.tmpl"),
			[]byte("A={{ .a }}"),
			0644,
		)
		require.NoError(t, err)

		err = os.WriteFile(
			filepath.Join(dir, "b.tmpl"),
			[]byte("B={{ .b }}"),
			0644,
		)
		require.NoError(t, err)

		fsys := os.DirFS(dir)
		tmpl := &template.Template{
			Name: "root",
			Files: []template.File{
				{
					Src:  "a.tmpl",
					Dest: "{{ .name }}/a.txt",
				},
				{
					Src:  "b.tmpl",
					Dest: "{{ .name }}/b.txt",
				},
			},
		}

		node := &template.Node{
			ID:       "0",
			Template: tmpl,
			FS:       fsys,
			Path:     ".",
		}

		out, err := r.RenderAll(
			node,
			template.RenderContexts{
				"0": testutil.Context(map[string]any{
					"name": "output",
					"a":    1,
					"b":    2,
				}),
			},
		)

		require.NoError(t, err)
		assert.Len(t, out.Files["0"], 2)

		resMap := make(map[string]string)
		for _, f := range out.Files["0"] {
			resMap[f.Path] = string(f.Content)
		}

		assert.Equal(t, "A=1", resMap["output/a.txt"])
		assert.Equal(t, "B=2", resMap["output/b.txt"])
	})
}
