package cli

import (
	"github.com/spf13/cobra"

	"composepack/internal/app"
)

// NewLogsCommand defines `composepack logs` placeholder logic.
func NewLogsCommand(application *app.Application) *cobra.Command {
	var (
		follow bool
		tail   int
	)

	cmd := &cobra.Command{
		Use:   "logs <release>",
		Short: "Follow docker compose logs for a release",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			releaseDir, err := cmd.Flags().GetString("release-dir")
			if err != nil {
				return err
			}

			opts := app.LogsOptions{
				ReleaseName:    args[0],
				RuntimeBaseDir: releaseDir,
				Follow:         follow,
				Tail:           tail,
			}

			return application.StreamLogs(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "stream logs")
	cmd.Flags().IntVar(&tail, "tail", 100, "lines to show from the end of the logs")

	return cmd
}
