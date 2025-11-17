package cli

import (
	"github.com/spf13/cobra"

	"composepack/internal/app"
)

// NewUpCommand wires the `composepack up` command skeleton.
func NewUpCommand(application *app.Application) *cobra.Command {
	var (
		valueFiles []string
		setValues  []string
		chartSrc   string
		detach     bool
	)

	cmd := &cobra.Command{
		Use:   "up <release>",
		Short: "Render and run docker compose up for a release",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			overrides, err := parseSetFlags(setValues)
			if err != nil {
				return err
			}

			releaseDir, err := cmd.Flags().GetString("release-dir")
			if err != nil {
				return err
			}

			opts := app.UpOptions{
				RenderOptions: app.RenderOptions{
					ReleaseName:    args[0],
					ChartSource:    chartSrc,
					ValueFiles:     append([]string{}, valueFiles...),
					SetValues:      overrides,
					RuntimeBaseDir: releaseDir,
				},
				Detach: detach,
			}

			return application.UpRelease(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&chartSrc, "chart", "", "optional chart directory or archive")
	cmd.Flags().StringArrayVarP(&valueFiles, "values", "f", nil, "values files to include")
	cmd.Flags().StringArrayVar(&setValues, "set", nil, "direct values to set")
	cmd.Flags().BoolVarP(&detach, "detach", "d", false, "pass --detach to docker compose up")

	return cmd
}
