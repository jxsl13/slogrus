package slogrus_test

import (
	"os"
	"time"

	"log/slog"

	"github.com/jxsl13/slogrus"
	"github.com/sirupsen/logrus"
)

func ExampleLogger() {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)

	// local variable
	slog := slog.New(slogrus.NewHandler(log, nil))

	slog.Debug("debug message", "key", "value")
	slog.Info("info message", "key", "value")
	slog.Warn("warn message", "key", "value")
	slog.Error("error message", "key", "value")
}

func ExampleDefaultLogger() {
	// global package
	slog.SetDefault(slog.New(slogrus.NewHandler(logrus.StandardLogger(), nil)))

	slog.Debug("debug message", "key", "value")
	slog.Info("info message", "key", "value")
	slog.Warn("warn message", "key", "value")
	slog.Error("error message", "key", "value")
}

func ExampleJSONLogger() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})

	// global package
	slog.SetDefault(slog.New(slogrus.NewHandler(log, nil)))

	slog.Debug("debug message", "key", "value")
	slog.Info("info message", "key", "value")
	slog.Warn("warn message", "key", "value")
	slog.Error("error message", "key", "value")
}
