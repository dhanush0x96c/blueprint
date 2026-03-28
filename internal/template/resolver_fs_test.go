package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFSResolver_Exists(t *testing.T) {
	base := t.TempDir()
	resolver := NewFSResolver(os.DirFS(base))

	templateName := "exists"
	dir := filepath.Join(base, templateName)
	writeTemplate(t, dir, validProjectTemplate)

	require.True(t, resolver.Exists(templateName))
	require.False(t, resolver.Exists("missing"))
}

func TestFSResolver_Discover(t *testing.T) {
	base := t.TempDir()
	resolver := NewFSResolver(os.DirFS(base))

	writeTemplate(t, filepath.Join(base, "projects", "go-cli"), validProjectTemplate)
	writeTemplate(t, filepath.Join(base, "features", "testing"), validFeatureTemplate)
	writeTemplate(t, filepath.Join(base, "broken"), invalidTemplate)

	templates, err := resolver.Discover()
	require.NoError(t, err)
	require.Len(t, templates, 2)
	require.Contains(t, templates, "projects/go-cli")
	require.Contains(t, templates, "features/testing")
	require.Equal(t, "go-cli", templates["projects/go-cli"].Name)
	require.Equal(t, "testing", templates["features/testing"].Name)
}
