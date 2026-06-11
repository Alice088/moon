// Package logger configures the zerolog global logger.
package logger

import (
	"go-template/internal/config"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup applies the log level and output format to the global logger.
func Setup(cfg config.LogConfig) {
	switch cfg.Level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if cfg.Format == "pretty" {
		log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = "2006-01-02 15:04:05"
		}))
	} else {
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

	log.Debug().Str("level", cfg.Level).Str("format", cfg.Format).Msg("logger configured")
}
