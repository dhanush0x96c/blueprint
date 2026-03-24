package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dhanush0x96c/blueprint/internal/app"
	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/fatih/color"
)

// TemplateListEntry represents a single template in the list output.
type TemplateListEntry struct {
	Name        string
	Type        template.Type
	Description string
}

// TemplateListGroup represents a group of templates from a single source.
type TemplateListGroup struct {
	Source  string // "BUILTIN" or "USER"
	Entries []TemplateListEntry
}

const (
	columnPadding = 2
)

var (
	sourceColor = color.New(color.FgHiWhite, color.Bold, color.Underline)
	nameColor   = color.New(color.FgBlue, color.Bold)
	descColor   = color.New(color.Faint)

	typeColors = map[template.Type]*color.Color{
		template.TypeProject:   color.New(color.FgYellow),
		template.TypeFeature:   color.New(color.FgCyan),
		template.TypeComponent: color.New(color.FgMagenta),
	}
)

// RenderTemplateList renders grouped template listings to stdout.
// When showType is true, the TYPE column is displayed in table output.
func RenderTemplateList(groups []TemplateListGroup, short, showType bool) {
	w := os.Stdout

	if short {
		renderShort(w, groups)
		return
	}

	renderTable(w, groups, showType)
}

func renderShort(w io.Writer, groups []TemplateListGroup) {
	for _, g := range groups {
		for _, e := range g.Entries {
			writeln(w, e.Name)
		}
	}
}

func renderTable(w io.Writer, groups []TemplateListGroup, showType bool) {
	nameWidth, typeWidth := calculateColumnWidths(groups)

	for i, g := range groups {
		if len(g.Entries) == 0 {
			continue
		}

		if i > 0 {
			writeln(w, "")
		}

		sourceColor.Fprintln(w, g.Source)

		for _, e := range g.Entries {
			fmt.Fprint(w, "  ")
			nameColor.Fprintf(w, "%-*s ", nameWidth, e.Name)
			if showType {
				colorForType(e.Type).Fprintf(w, "%-*s ", typeWidth, e.Type)
			}
			descColor.Fprintln(w, e.Description)
		}
	}
}

func calculateColumnWidths(groups []TemplateListGroup) (nameWidth, typeWidth int) {
	for _, g := range groups {
		for _, e := range g.Entries {
			if len(e.Name) > nameWidth {
				nameWidth = len(e.Name)
			}
			if len(e.Type) > typeWidth {
				typeWidth = len(e.Type)
			}
		}
	}
	nameWidth += columnPadding
	typeWidth += columnPadding
	return
}

func colorForType(t template.Type) *color.Color {
	if c, ok := typeColors[t]; ok {
		return c
	}
	return color.New(color.FgWhite)
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
		return "", &app.InvalidTemplateTypeError{Type: arg}
	}
}
