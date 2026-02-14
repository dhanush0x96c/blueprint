package cmd

import (
	"fmt"

	"github.com/dhanush0x96c/blueprint/internal/app"
	"github.com/dhanush0x96c/blueprint/internal/config"
	"github.com/dhanush0x96c/blueprint/internal/ui"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cfgLoader := config.Loader{
		EnvPrefix: "BLUEPRINT",
		CLIArgs:   map[string]string{},
	}
	var appCtx = new(app.Context)
	var options = app.Options{}

	cmd := &cobra.Command{
		Use:           "blueprint",
		Aliases:       []string{"bp"},
		Short:         "Universal project scaffolding",
		Long:          "Blueprint scaffolds projects from composable templates.",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			ctx := app.NewContext(cfg, options)
			*appCtx = *ctx

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(
		&cfgLoader.ConfigFile,
		"config",
		"",
		fmt.Sprintf("config file (default is %s)", config.DefaultPathUsage()),
	)

	cmd.PersistentFlags().BoolVarP(
		&options.Verbose,
		"verbose",
		"v",
		false,
		"Enable verbose output",
	)

	cmd.PersistentFlags().BoolVar(
		&options.DryRun,
		"dry-run",
		false,
		"Preview actions without writing files",
	)

	cmd.AddCommand(NewInitCommand(appCtx))
	cmd.AddCommand(NewVersionCommand(appCtx))

	return cmd
}

func Execute() int {
	if err := NewRootCmd().Execute(); err != nil {
		ui.RenderError(err)
		return ui.ExitCode(err)
	}
	return 0
}
