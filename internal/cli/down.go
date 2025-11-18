package cli

import (
	"github.com/spf13/cobra"

	"composepack/internal/app"
)

// NewDownCommand defines `composepack down` placeholder logic.
func NewDownCommand(application *app.Application) *cobra.Command {
	var (
		removeVolumes bool
		runtimeDir    string
	)

	cmd := &cobra.Command{
		Use:   "down <release>",
		Short: "Run docker compose down for a release runtime",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			releaseDir, err := cmd.Flags().GetString("release-dir")
			if err != nil {
				return err
			}

			opts := app.DownOptions{
				ReleaseName:    args[0],
				RuntimeBaseDir: releaseDir,
				RuntimePath:    runtimeDir,
				RemoveVolumes:  removeVolumes,
			}

			return application.DownRelease(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&removeVolumes, "volumes", false, "include volumes when bringing the release down")
	cmd.Flags().StringVar(&runtimeDir, "runtime-dir", "", "path to release directory (overrides --release-dir)")

	return cmd
}
