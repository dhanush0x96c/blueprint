package cmd

import (
	"io/fs"
	"sort"
	"strings"

	"github.com/dhanush0x96c/blueprint/internal/app"
	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/ui"
	"github.com/spf13/cobra"
)

func NewListCmd(appCtx *app.Context) *cobra.Command {
	var (
		source string
		short  bool
		tags   []string
	)

	cmd := &cobra.Command{
		Use:   "list [projects|features|components]",
		Short: "List available templates",
		Long:  "List available templates, optionally filtered by type, source, and tags.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var filterType template.Type
			showType := len(args) == 0
			if !showType {
				t, err := ui.ValidTemplateTypeArg(args[0])
				if err != nil {
					return err
				}
				filterType = t
			}

			groups, err := discoverTemplates(appCtx, filterType, source, tags)
			if err != nil {
				return err
			}

			ui.RenderTemplateList(groups, short, showType)
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

	cmd.Flags().StringSliceVar(
		&tags,
		"tags",
		nil,
		"Filter by tags (comma-separated). Matches templates that contain ANY of the specified tags.",
	)

	return cmd
}

func discoverTemplates(
	appCtx *app.Context,
	filterType template.Type,
	source string,
	tags []string,
) ([]ui.TemplateListGroup, error) {
	var groups []ui.TemplateListGroup

	if source == "" || source == "builtin" {
		entries, err := discoverFromFS(appCtx.BuiltinFS, filterType, tags)
		if err != nil {
			return nil, err
		}
		groups = append(groups, ui.TemplateListGroup{
			Source:  "BUILTIN",
			Entries: entries,
		})
	}

	if source == "" || source == "user" {
		entries, err := discoverFromFS(appCtx.LocalFS, filterType, tags)
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

func discoverFromFS(fsys fs.FS, filterType template.Type, filterTags []string) ([]ui.TemplateListEntry, error) {
	loader := template.NewLoader(fsys)
	templates, err := loader.DiscoverAll(filterType)
	if err != nil {
		return nil, err
	}

	entries := make([]ui.TemplateListEntry, 0, len(templates))
	for _, tmpl := range templates {
		if len(filterTags) > 0 && !matchesAnyTag(tmpl, filterTags) {
			continue
		}

		entries = append(entries, ui.TemplateListEntry{
			Name:        tmpl.Name,
			Type:        tmpl.Type,
			Description: tmpl.Description,
		})
	}

	typeOrder := map[template.Type]int{
		template.TypeProject:   0,
		template.TypeFeature:   1,
		template.TypeComponent: 2,
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Type != entries[j].Type {
			return typeOrder[entries[i].Type] < typeOrder[entries[j].Type]
		}
		return entries[i].Name < entries[j].Name
	})

	return entries, nil
}

// matchesAnyTag returns true if the template has at least one of the filter tags
func matchesAnyTag(tmpl *template.Template, filterTags []string) bool {
	if len(tmpl.Tags) == 0 {
		return false
	}

	tagSet := make(map[string]struct{}, len(tmpl.Tags))
	for _, t := range tmpl.Tags {
		tagSet[strings.ToLower(t)] = struct{}{}
	}

	for _, ft := range filterTags {
		if _, ok := tagSet[strings.ToLower(ft)]; ok {
			return true
		}
	}

	return false
}
