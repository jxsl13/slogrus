package slogrus_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/jxsl13/slogrus"
	"github.com/sirupsen/logrus"
)

func TestLogrusHandler(t *testing.T) {
	var bufA bytes.Buffer
	var bufB bytes.Buffer

	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	l.SetLevel(logrus.DebugLevel)
	l.SetOutput(io.MultiWriter(&bufA))

	opts := &slog.HandlerOptions{
		AddSource: true,
	}

	h := slogrus.NewHandler(l, opts)

	log := slog.New(h)

	log.WithGroup("group").Error("group error", "groupKey", slog.GroupValue(slog.String("groupkey1", "groupValue1"), slog.String("groupKey2", "groupValue2")))

	log = slog.New(slog.NewJSONHandler(io.MultiWriter(&bufB), opts))
	log.WithGroup("group").Error("group error", "groupKey", slog.GroupValue(slog.String("groupkey1", "groupValue1"), slog.String("groupKey2", "groupValue2")))

	var a map[string]any
	var b map[string]any

	json.Unmarshal(bufA.Bytes(), &a)
	json.Unmarshal(bufB.Bytes(), &b)

	delete(a, "level") // lowercase
	delete(b, "level") // uppercase

	delete(a, "time") // differs
	delete(b, "time") // differes

	delete(a[slog.SourceKey].(map[string]any), "line") // also differs
	delete(b[slog.SourceKey].(map[string]any), "line") // also differs

	if !reflect.DeepEqual(a, b) {
		t.Fatalf("expected %v, got %v", a, b)
	}
}
