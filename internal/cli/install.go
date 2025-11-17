package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"composepack/internal/app"
)

// NewInstallCommand returns the `composepack install` Cobra command skeleton.
func NewInstallCommand(application *app.Application) *cobra.Command {
	var (
		releaseName string
		valueFiles  []string
		setValues   []string
		autoStart   bool
	)

	cmd := &cobra.Command{
		Use:   "install <chart>",
		Short: "Install a chart into a named release runtime",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("chart source path must be specified")
			}
			if releaseName == "" {
				return fmt.Errorf("--name is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			chartSource := args[0]

			overrides, err := parseSetFlags(setValues)
			if err != nil {
				return err
			}

			releaseDir, err := cmd.Flags().GetString("release-dir")
			if err != nil {
				return err
			}

			opts := app.InstallOptions{
				RenderOptions: app.RenderOptions{
					ReleaseName:    releaseName,
					ChartSource:    chartSource,
					ValueFiles:     append([]string{}, valueFiles...),
					SetValues:      overrides,
					RuntimeBaseDir: releaseDir,
				},
				AutoStart: autoStart,
			}

			return application.InstallRelease(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&releaseName, "name", "", "release name to use for the installation")
	cmd.Flags().StringArrayVarP(&valueFiles, "values", "f", nil, "values files to include (can specify multiple)")
	cmd.Flags().StringArrayVar(&setValues, "set", nil, "direct value overrides (key=value)")
	cmd.Flags().BoolVar(&autoStart, "auto-start", false, "run docker compose up after installation")

	return cmd
}
