package logger

import (
	"io"
	"log/slog"
)

func GetLogger(out io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, nil))
}
