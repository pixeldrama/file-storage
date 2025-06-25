package loginit

import (
	"log/slog"
	"os"
	"strings"

	"github.com/benjamin/file-storage-go/pkg/ecsslog"
)

// InitLogger initializes the global slog logger based on the LOG_FORMAT environment variable.
// If LOG_FORMAT=ecs, it uses the ECS handler; otherwise, it uses the default text handler.
// Returns the configured logger.
func InitLogger() *slog.Logger {
	format := strings.ToLower(os.Getenv("LOG_FORMAT"))
	var handler slog.Handler

	if format == "ecs" {
		handler = ecsslog.NewECSHandler(os.Stdout, ecsslog.Config{})
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
