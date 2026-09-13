package app

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/natefinch/lumberjack.v2"
)

// IsDevEnv controls whether logs go to stdout (true) or to a file (false).
var IsDevEnv bool = false

// InitLogger sets up the slog logger with log rotation.
func InitLogger() error {
	var handler slog.Handler

	if IsDevEnv {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		var logDir string
		switch runtime.GOOS {
		case "windows":
			logDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "LanShare", "logs")
		case "darwin":
			home, _ := os.UserHomeDir()
			logDir = filepath.Join(home, "Library", "Logs", "LanShare")
		default: // linux and others
			home, _ := os.UserHomeDir()
			logDir = filepath.Join(home, ".local", "state", "lanshare", "logs")
		}

		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		logFile := filepath.Join(logDir, "lanshare.log")

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
