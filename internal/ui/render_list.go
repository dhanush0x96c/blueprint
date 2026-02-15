package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// TemplateListEntry represents a single template in the list output.
type TemplateListEntry struct {
	Name        string
	Description string
}

// TemplateListGroup represents a group of templates from a single source.
type TemplateListGroup struct {
	Source  string // "BUILTIN" or "USER"
	Entries []TemplateListEntry
}

const (
	listNameWidth = 25
	listSeparator = "─────────────────────────────────────────────────────────────────────"
)

// RenderTemplateList renders grouped template listings to stdout.
func RenderTemplateList(groups []TemplateListGroup, short bool) {
	w := os.Stdout

	if short {
		renderShort(w, groups)
		return
	}

	renderTable(w, groups)
}

func renderShort(w io.Writer, groups []TemplateListGroup) {
	for _, g := range groups {
		for _, e := range g.Entries {
			writeln(w, e.Name)
		}
	}
}

func renderTable(w io.Writer, groups []TemplateListGroup) {
	for i, g := range groups {
		if len(g.Entries) == 0 {
			continue
		}

		if i > 0 {
			writeln(w, "")
		}

		write(w, "%s TEMPLATES\n", g.Source)
		writeln(w, listSeparator)
		write(w, "%-*s %s\n", listNameWidth, "NAME", "DESCRIPTION")
		writeln(w, listSeparator)

		for _, e := range g.Entries {
			write(w, "%-*s %s\n", listNameWidth, e.Name, e.Description)
		}
	}
}

// ValidTemplateTypeArg checks if the given argument is a valid template type filter.
func ValidTemplateTypeArg(arg string) (template.Type, error) {
	switch strings.ToLower(arg) {
	case "projects":
		return template.TypeProject, nil
	case "features":
		return template.TypeFeature, nil
	case "components":
		return template.TypeComponent, nil
	default:
		return "", fmt.Errorf("invalid template type %q: expected projects, features, or components", arg)
	}
}
