//go:build wireinject

//go:generate go run -mod=mod github.com/google/wire/cmd/wire

package di

import (
	"github.com/google/wire"

	"composepack/internal/app"
	"composepack/internal/core/chart"
	"composepack/internal/infra/config"
	"composepack/internal/infra/logging"
)

// provideConfig supplies the base configuration used across the application.
func provideConfig() config.Config {
	return config.Default()
}

// provideLogger supplies the root logging implementation.
func provideLogger() logging.Logger {
	return logging.NewZerolog(nil)
}

// InitializeApplication wires together all dependencies via Google Wire.
func InitializeApplication() (*app.Application, error) {
	wire.Build(
		provideConfig,
		provideLogger,
		chart.ProviderSet,
		app.NewRuntime,
		app.NewApplication,
	)
	return nil, nil
}
