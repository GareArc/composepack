package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"composepack/internal/version"
)

// NewVersionCommand prints the current ComposePack CLI version.
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the composepack CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "composepack %s\n", version.Version)
		},
	}
	return cmd
}
