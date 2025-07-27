// internal/util/logger.go
package Util

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func init() {
	logFile, err := os.OpenFile("scraper.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}

	Logger = slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}
