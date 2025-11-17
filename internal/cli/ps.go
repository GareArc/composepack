package cli

import (
	"github.com/spf13/cobra"

	"composepack/internal/app"
)

// NewPSCommand defines `composepack ps` placeholder logic.
func NewPSCommand(application *app.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ps <release>",
		Short: "Display docker compose ps for a release",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			releaseDir, err := cmd.Flags().GetString("release-dir")
			if err != nil {
				return err
			}

			opts := app.PSOptions{
				ReleaseName:    args[0],
				RuntimeBaseDir: releaseDir,
			}

			return application.ShowStatus(cmd.Context(), opts)
		},
	}

	return cmd
}
