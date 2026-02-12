package cmd

import (
	"fmt"

	"github.com/dhanush0x96c/blueprint/internal/app"
	"github.com/dhanush0x96c/blueprint/internal/scaffold"
	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/ui"
	"github.com/spf13/cobra"
)

func NewInitCommand(appCtx *app.Context) *cobra.Command {
	var (
		outputDir string
		force     bool
		yes       bool
	)

	cmd := &cobra.Command{
		Use:   "init <template>",
		Short: "Initialize a new project",
		Long:  `Initialize a new project from a template.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]

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

	cmd.Flags().StringVarP(
		&outputDir,
		"output",
		"o",
		"",
		"Output directory",
	)

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

	return cmd
}
