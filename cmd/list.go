package cmd

import (
	"io/fs"
	"sort"

	"github.com/dhanush0x96c/blueprint/internal/app"
	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/ui"
	"github.com/spf13/cobra"
)

func NewListCmd(appCtx *app.Context) *cobra.Command {
	var (
		source string
		short  bool
	)

	cmd := &cobra.Command{
		Use:   "list [projects|features|components]",
		Short: "List available templates",
		Long:  "List available templates, optionally filtered by type and source.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var filterType template.Type
			if len(args) > 0 {
				t, err := ui.ValidTemplateTypeArg(args[0])
				if err != nil {
					return err
				}
				filterType = t
			}

			groups, err := discoverTemplates(appCtx, filterType, source)
			if err != nil {
				return err
			}

			ui.RenderTemplateList(groups, short)
			return nil
		},
	}

	cmd.Flags().StringVar(
		&source,
		"source",
		"",
		"Filter by source: builtin, user (default: all)",
	)

	cmd.Flags().BoolVar(
		&short,
		"short",
		false,
		"Show compact output (name only)",
	)

	return cmd
}

func discoverTemplates(appCtx *app.Context, filterType template.Type, source string) ([]ui.TemplateListGroup, error) {
	var groups []ui.TemplateListGroup

	if source == "" || source == "builtin" {
		entries, err := discoverFromFS(appCtx.BuiltinFS, filterType)
		if err != nil {
			return nil, err
		}
		groups = append(groups, ui.TemplateListGroup{
			Source:  "BUILTIN",
			Entries: entries,
		})
	}

	if source == "" || source == "user" {
		entries, err := discoverFromFS(appCtx.LocalFS, filterType)
		if err != nil {
			// User template dir may not exist; treat as empty
			entries = nil
		}
		groups = append(groups, ui.TemplateListGroup{
			Source:  "USER",
			Entries: entries,
		})
	}

	return groups, nil
}

func discoverFromFS(fsys fs.FS, filterType template.Type) ([]ui.TemplateListEntry, error) {
	loader := template.NewLoader(fsys)
	templates, err := loader.DiscoverAll(filterType)
	if err != nil {
		return nil, err
	}

	entries := make([]ui.TemplateListEntry, 0, len(templates))
	for _, tmpl := range templates {
		entries = append(entries, ui.TemplateListEntry{
			Name:        tmpl.Name,
			Description: tmpl.Description,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries, nil
}
