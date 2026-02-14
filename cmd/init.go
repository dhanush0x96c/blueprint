package cmd

import (
	"fmt"
	"strings"

	"github.com/dhanush0x96c/blueprint/internal/app"
	"github.com/dhanush0x96c/blueprint/internal/scaffold"
	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/ui"
	"github.com/spf13/cobra"
)

func NewInitCommand(appCtx *app.Context) *cobra.Command {
	var (
		force    bool
		yes      bool
		varFlags []string
	)

	cmd := &cobra.Command{
		Use:   "init <template> [output-dir]",
		Short: "Initialize a new project",
		Long:  `Initialize a new project from a template.`,
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]

			var outputDir string
			if len(args) > 1 {
				outputDir = args[1]
			}

			vars, err := parseVarFlags(varFlags)
			if err != nil {
				return err
			}

			resolved, err := appCtx.Resolver.Resolve(appCtx, app.TemplateRef{
				Name: templateName,
				Type: template.TypeProject,
			})

			if err != nil {
				return fmt.Errorf("failed to resolve template %s: %w", templateName, err)
			}

			scaffolder := scaffold.NewScaffolder(resolved.FS)
			result, err := scaffolder.Scaffold(scaffold.Options{
				TemplatePath: resolved.Path,
				OutputDir:    outputDir,
				Variables:    vars,
				Interactive:  !yes,
				Overwrite:    force,
			})

			if err != nil {
				return fmt.Errorf("init template %q: %w", templateName, err)
			}

			ui.RenderResult(result)

			return nil
		},
	}

	cmd.Flags().BoolVarP(
		&force,
		"force",
		"f",
		false,
		"Overwrite existing files if they exist",
	)

	cmd.Flags().BoolVarP(
		&yes,
		"yes",
		"y",
		false,
		"Accept defaults and disable prompts",
	)

	cmd.Flags().StringArrayVar(
		&varFlags,
		"var",
		nil,
		`Set a template variable (format: key=value)`,
	)

	return cmd
}

func parseVarFlags(flags []string) (map[string]any, error) {
	if len(flags) == 0 {
		return nil, nil
	}

	vars := make(map[string]any, len(flags))
	for _, f := range flags {
		key, value, ok := strings.Cut(f, "=")
		if !ok {
			return nil, fmt.Errorf("invalid variable format %q: expected key=value", f)
		}
		vars[key] = value
	}
	return vars, nil
}
