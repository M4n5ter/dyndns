package log

import (
	"golang.org/x/exp/slog"
)

func LogPanic(logger *slog.Logger, msg string) {
	logger.Error(msg)
	panic(msg)
}
