package ecsslog

import (
	"context"
	"io"
	"log/slog"
	"os"
)

const (
	ecsVersion = "8.11.0"
	logger     = "log/slog"
)

const (
	ecsVersionKey = "ecs.version"

	timestampKey = "@timestamp"
	messageKey   = "message"
	logLevelKey  = "log.level"
	logLoggerKey = "log.logger"
)

type Handler struct {
	jsonHandler slog.Handler
	levelNamer  func(slog.Level) string
}

type Config struct {
	HandlerOptions slog.HandlerOptions

	// enables customizing of how log levels would look (INFO, info, INF, etc.)
	LevelNamer func(slog.Level) string
}

func NewECSHandler(writer io.Writer, config Config) *Handler {
	if config.LevelNamer == nil {
		config.LevelNamer = defaultLevelNamer
	}
	if writer == nil {
		writer = os.Stdout
	}
	config.HandlerOptions.ReplaceAttr = removeJsonHandlerAttrs

	return &Handler{
		jsonHandler: slog.NewJSONHandler(writer, &config.HandlerOptions),
		levelNamer:  config.LevelNamer,
	}
}

// slog.JsonHandler has opinions about some field names. This removes all of them, so we can later add ECS compliant ones.
func removeJsonHandlerAttrs(groups []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case "time", "msg", "source", "level":
		return slog.Attr{}
	default:
		return a
	}
}

func defaultLevelNamer(l slog.Level) string { return l.String() }

func (x *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return x.jsonHandler.Enabled(ctx, level)
}

func (x *Handler) Handle(ctx context.Context, record slog.Record) error {
	record.AddAttrs(
		slog.Time(timestampKey, record.Time),
		slog.String(messageKey, record.Message),
		slog.String(logLevelKey, x.levelNamer(record.Level)),
		slog.String(ecsVersionKey, ecsVersion),
		slog.String(logLoggerKey, logger),
	)
	return x.jsonHandler.Handle(ctx, record)
}

func (x *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{jsonHandler: x.jsonHandler.WithAttrs(attrs), levelNamer: x.levelNamer}
}

func (x *Handler) WithGroup(name string) slog.Handler {
	return &Handler{jsonHandler: x.jsonHandler.WithGroup(name), levelNamer: x.levelNamer}
}
