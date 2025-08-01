package helpers

import (
	"brevet-api/middlewares"
	"context"

	"github.com/sirupsen/logrus"
)

// LoggerFromCtx get logger context
func LoggerFromCtx(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	if logger, ok := ctx.Value(middlewares.LoggerKey()).(*logrus.Entry); ok {
		return logger
	}
	return logrus.NewEntry(logrus.StandardLogger())
}
