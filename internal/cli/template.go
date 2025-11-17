package cli

import (
	"github.com/spf13/cobra"

	"composepack/internal/app"
)

// NewTemplateCommand wires the `composepack template` command skeleton.
func NewTemplateCommand(application *app.Application) *cobra.Command {
	var (
		valueFiles []string
		setValues  []string
		chartSrc   string
	)

	cmd := &cobra.Command{
		Use:   "template <release>",
		Short: "Render a release runtime without invoking docker compose",
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

			opts := app.TemplateOptions{
				RenderOptions: app.RenderOptions{
					ReleaseName:    args[0],
					ChartSource:    chartSrc,
					ValueFiles:     append([]string{}, valueFiles...),
					SetValues:      overrides,
					RuntimeBaseDir: releaseDir,
				},
			}

			return application.TemplateRelease(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&chartSrc, "chart", "", "chart directory or archive to render")
	cmd.Flags().StringArrayVarP(&valueFiles, "values", "f", nil, "values files to include")
	cmd.Flags().StringArrayVar(&setValues, "set", nil, "direct values to set (key=value)")

	return cmd
}
