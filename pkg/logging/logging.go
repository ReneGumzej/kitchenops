// Package logging stellt das zentrale, strukturierte Logging der App bereit.
// Ausgabe erfolgt auf stdout (Docker-Prinzip: "docker logs" sammelt alles).
//
// Konfiguration über Umgebungsvariablen:
//
//	LOG_LEVEL   debug | info | warn | error   (Standard: info)
//	LOG_FORMAT  text | json                   (Standard: text)
package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Vorausinitialisierter Logger, damit Pakete auch ohne explizites Init
// (z.B. in Tests) gefahrlos loggen können.
var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// Init konfiguriert das Logging aus den Umgebungsvariablen und setzt es
// als slog-Standard.
func Init() *slog.Logger {
	var level slog.Level
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if strings.EqualFold(strings.TrimSpace(os.Getenv("LOG_FORMAT")), "json") {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger = slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// Logger liefert den aktiven Logger.
func Logger() *slog.Logger { return logger }

func Debug(msg string, args ...any) { logger.Debug(msg, args...) }
func Info(msg string, args ...any)  { logger.Info(msg, args...) }
func Warn(msg string, args ...any)  { logger.Warn(msg, args...) }
func Error(msg string, args ...any) { logger.Error(msg, args...) }

// Kürzen schneidet lange Werte für die Log-Ausgabe ab.
func Kürzen(wert string, max int) string {
	if len(wert) <= max {
		return wert
	}
	return wert[:max] + "…"
}
