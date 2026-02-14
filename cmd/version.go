package cmd

import (
	"fmt"

	"github.com/dhanush0x96c/blueprint/internal/app"
	"github.com/dhanush0x96c/blueprint/internal/version"
	"github.com/spf13/cobra"
)

func NewVersionCommand(appCtx *app.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print the version, commit hash, and build date of Blueprint.",
		Run: func(cmd *cobra.Command, args []string) {
			if appCtx.Options.Verbose {
				fmt.Printf("Blueprint %s\n", version.Version)
				fmt.Printf("Git Commit: %s\n", version.GitCommit)
				fmt.Printf("Build Date: %s\n", version.BuildDate)
			} else {
				fmt.Printf("Blueprint %s\n", version.Version)
			}
		},
	}
}
