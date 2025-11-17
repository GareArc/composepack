package cli

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"composepack/internal/scaffold"
)

// NewInitCommand scaffolds a ComposePack chart directory.
func NewInitCommand() *cobra.Command {
	var opts scaffold.Options

	cmd := &cobra.Command{
		Use:   "init <path>",
		Short: "Create a starter ComposePack chart at the given path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Path = args[0]
			if opts.Name == "" {
				opts.Name = filepath.Base(args[0])
			}
			return scaffold.CreateChart(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "chart name (defaults to directory name)")
	cmd.Flags().StringVar(&opts.Version, "version", "0.1.0", "chart version")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "overwrite existing files in target directory")

	return cmd
}
