package logging

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

// Logger describes the minimal logging interface used throughout the tool.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

// Nop satisfies Logger for tests and scaffolding.
type Nop struct{}

// Info implements Logger.
func (Nop) Info(msg string, args ...any) {}

// Warn implements Logger.
func (Nop) Warn(msg string, args ...any) {}

// Error implements Logger.
func (Nop) Error(msg string, args ...any) {}

// Debug implements Logger.
func (Nop) Debug(msg string, args ...any) {}

// ZerologLogger implements Logger using github.com/rs/zerolog.
type ZerologLogger struct {
	logger zerolog.Logger
}

// NewZerolog returns a Zerolog-backed Logger writing to the provided io.Writer.
func NewZerolog(writer io.Writer) Logger {
	if writer == nil {
		writer = os.Stderr
	}
	l := zerolog.New(writer).With().Timestamp().Logger()
	return ZerologLogger{logger: l}
}

// Info implements Logger.
func (z ZerologLogger) Info(msg string, args ...any) {
	z.logger.Info().Msgf(msg, args...)
}

// Warn implements Logger.
func (z ZerologLogger) Warn(msg string, args ...any) {
	z.logger.Warn().Msgf(msg, args...)
}

// Error implements Logger.
func (z ZerologLogger) Error(msg string, args ...any) {
	z.logger.Error().Msgf(msg, args...)
}

// Debug implements Logger.
func (z ZerologLogger) Debug(msg string, args ...any) {
	z.logger.Debug().Msgf(msg, args...)
}
