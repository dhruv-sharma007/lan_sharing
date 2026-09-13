package app

import (
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// IsDevEnv controls whether logs go to stdout (true) or to a file (false).
var IsDevEnv bool = true

// InitLogger sets up the slog logger with log rotation.
func InitLogger() error {
	var handler slog.Handler

	if IsDevEnv {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		// config.json is loaded from the working directory, so keep the log beside it.
		logFile := "lanshare.log"

		rotator := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
			Compress:   true,
		}

		// Use slog with JSON or text format
		handler = slog.NewTextHandler(rotator, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Logger initialized", "is_dev", IsDevEnv)
	return nil
}
