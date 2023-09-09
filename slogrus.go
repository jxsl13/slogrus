package slogrus

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strconv"

	"github.com/sirupsen/logrus"
)

// NewHandler uses a logrus logger to wrap an slog
func NewHandler(log *logrus.Logger, opts *slog.HandlerOptions) slog.Handler {
	var opt slog.HandlerOptions
	if opts != nil {
		opt = *opts
	}
	// must not change logrus' internal config because we are not the owner of the logger
	opt.Level = fromLogrusLogLevel(log.Level)
	opt.AddSource = log.ReportCaller || opt.AddSource

	return &logrusHandler{
		opt:        opt,
		log:        log,
		attributes: make(map[string]any),
		groups:     make([]string, 0),
	}
}

type logrusHandler struct {
	opt        slog.HandlerOptions
	log        *logrus.Logger
	attributes map[string]any
	groups     []string
}

func (l *logrusHandler) Enabled(_ context.Context, level slog.Level) bool {
	return l.log.IsLevelEnabled(toLogrusLevel(level))
}

func (l *logrusHandler) Handle(ctx context.Context, record slog.Record) error {
	e := l.log.WithTime(record.Time)

	add := record.NumAttrs()
	if l.opt.AddSource {
		add++
	}
	attr := cloneMap(l.attributes, add)

	record.Attrs(func(a slog.Attr) bool {
		if l.opt.ReplaceAttr != nil {
			a = l.opt.ReplaceAttr(l.groups, a)

			attr[a.Key] = toAny(a.Value)
		} else {

			attr[a.Key] = toAny(a.Value)
		}
		return true
	})

	fields := groupTree(l.groups, attr)

	// add keys on root level, not on group level
	if l.opt.AddSource {
		frames := runtime.CallersFrames([]uintptr{record.PC})
		f, _ := frames.Next()
		fields[slog.SourceKey] = map[string]any{
			"file":     f.File,
			"line":     strconv.Itoa(f.Line),
			"function": f.Function,
		}
	}

	e.WithFields(fields).WithContext(ctx).Log(toLogrusLevel(record.Level), record.Message)
	return nil
}

func (l *logrusHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	attributes := make(map[string]any, len(l.attributes)+len(attrs))
	for k, v := range l.attributes {
		attributes[k] = v
	}
	for _, attr := range attrs {
		if l.opt.ReplaceAttr != nil {
			attr = l.opt.ReplaceAttr(l.groups, attr)
			attributes[attr.Key] = toAny(attr.Value)
		} else {
			attributes[attr.Key] = toAny(attr.Value)
		}
	}

	return &logrusHandler{
		opt:        l.opt,
		log:        l.log,
		attributes: attributes,
		groups:     cloneSlice(l.groups),
	}
}

func (l *logrusHandler) WithGroup(name string) slog.Handler {
	groups := make([]string, len(l.groups)+1)
	copy(groups, l.groups)
	groups[len(l.groups)] = name

	return &logrusHandler{
		opt:        l.opt,
		log:        l.log,
		attributes: cloneMap(l.attributes),
		groups:     groups,
	}
}

func cloneMap(m map[string]any, addsize ...int) map[string]any {
	if m == nil {
		return nil
	}

	add := 0
	if len(addsize) > 0 && addsize[0] > 0 {
		add = addsize[0]
	}

	m2 := make(map[string]any, len(m)+add)
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

func cloneSlice(ss []string) []string {
	ss2 := make([]string, len(ss))
	copy(ss2, ss)
	return ss2
}

func groupTree(groups []string, attr map[string]any) map[string]any {
	if len(groups) == 0 {
		return attr
	}

	root := make(map[string]any, 1)
	walker := root
	for idx, group := range groups {

		if idx == len(groups)-1 {
			walker[group] = attr
		} else {
			m := make(map[string]any, 1)
			walker[group] = m
			walker = m
		}
	}

	return root
}

func toAny(v slog.Value) any {
	switch v.Kind() {
	case slog.KindBool:
		return v.Bool()
	case slog.KindInt64:
		return v.Int64()
	case slog.KindString:
		return v.String()
	case slog.KindDuration:
		return v.Duration()
	case slog.KindTime:
		return v.Time()
	case slog.KindUint64:
		return v.Uint64()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindGroup:
		attrs := v.Group()
		m := make(map[string]any, len(attrs))
		for _, attr := range attrs {
			m[attr.Key] = toAny(attr.Value)
		}
		return m
	case slog.KindAny:
		return v.Any()
	case slog.KindLogValuer:
		return toAny(v.LogValuer().LogValue())
	default:
		panic(fmt.Sprintf("bad kind: %s", v.Kind()))
	}
}

func addKeyValue(m map[string]any, key string, value any) map[string]any {
	m2 := make(map[string]any, len(m)+1)
	for k, v := range m {
		m2[k] = v
	}
	m2[key] = value
	return m2
}

func toLogrusLevel(level slog.Level) logrus.Level {
	switch level {
	case slog.LevelDebug:
		return logrus.DebugLevel
	case slog.LevelInfo:
		return logrus.InfoLevel
	case slog.LevelWarn:
		return logrus.WarnLevel
	case slog.LevelError:
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

func fromLogrusLogLevel(level logrus.Level) slog.Level {
	switch level {
	case logrus.DebugLevel, logrus.TraceLevel:
		return slog.LevelDebug
	case logrus.InfoLevel:
		return slog.LevelInfo
	case logrus.WarnLevel:
		return slog.LevelWarn
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
