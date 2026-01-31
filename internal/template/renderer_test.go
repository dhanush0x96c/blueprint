package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to create a renderer with a temp dir
func newTestRenderer(t *testing.T) (*Renderer, string) {
	t.Helper()

	dir := t.TempDir()
	r := NewRenderer(os.DirFS(dir))

	return r, dir
}

// helper context
func testContext(vars map[string]any) *Context {
	return &Context{
		Variables: vars,
	}
}

func TestRenderString_Simple(t *testing.T) {
	r, _ := newTestRenderer(t)

	out, err := r.RenderString(
		"Hello {{ .name }}",
		testContext(map[string]any{
			"name": "Blueprint",
		}),
		"test",
	)

	require.NoError(t, err)
	assert.Equal(t, "Hello Blueprint", out)
}

func TestRenderString_WithDefaultFuncs(t *testing.T) {
	r, _ := newTestRenderer(t)

	out, err := r.RenderString(
		"{{ .name | toUpper }}",
		testContext(map[string]any{
			"name": "blueprint",
		}),
		"test",
	)

	require.NoError(t, err)
	assert.Equal(t, "BLUEPRINT", out)
}

func TestRenderString_ParseError(t *testing.T) {
	r, _ := newTestRenderer(t)

	_, err := r.RenderString(
		"{{ .name ",
		testContext(map[string]any{
			"name": "oops",
		}),
		"broken",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse template")
}

func TestRenderString_ExecutionError(t *testing.T) {
	r, _ := newTestRenderer(t)

	// toInt returns (int, error) → template execution error
	_, err := r.RenderString(
		"{{ toInt .value }}",
		testContext(map[string]any{
			"value": "not-a-number",
		}),
		"exec-error",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute template")
}

func TestRender_File(t *testing.T) {
	r, dir := newTestRenderer(t)

	path := filepath.Join(dir, "hello.tmpl")
	err := os.WriteFile(path, []byte("Hi {{ .name }}"), 0644)
	require.NoError(t, err)

	out, err := r.Render(
		"hello.tmpl",
		testContext(map[string]any{
			"name": "World",
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, "Hi World", out)
}

func TestRender_FileNotFound(t *testing.T) {
	r, _ := newTestRenderer(t)

	_, err := r.Render(
		"missing.tmpl",
		testContext(map[string]any{}),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read template file")
}

func TestRenderPath(t *testing.T) {
	r, _ := newTestRenderer(t)

	out, err := r.RenderPath(
		"{{ .pkg }}/main.go",
		testContext(map[string]any{
			"pkg": "myapp",
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, "myapp/main.go", out)
}

func TestAddFunc_CustomFunction(t *testing.T) {
	r, _ := newTestRenderer(t)

	r.AddFunc("shout", func(s string) string {
		return s + "!!!"
	})

	out, err := r.RenderString(
		"{{ shout .msg }}",
		testContext(map[string]any{
			"msg": "hey",
		}),
		"custom-func",
	)

	require.NoError(t, err)
	assert.Equal(t, "hey!!!", out)
}

func TestRenderAll(t *testing.T) {
	r, dir := newTestRenderer(t)

	// create template files
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

	tmpl := &Template{
		Files: []File{
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

	out, err := r.RenderAll(
		tmpl,
		testContext(map[string]any{
			"name": "output",
			"a":    1,
			"b":    2,
		}),
	)

	require.NoError(t, err)
	assert.Len(t, out, 2)

	assert.Equal(t, "A=1", out["output/a.txt"])
	assert.Equal(t, "B=2", out["output/b.txt"])
}
