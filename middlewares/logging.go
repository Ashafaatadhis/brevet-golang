package middlewares

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ctxKey string

const (
	loggerKey  ctxKey = "logger"
	traceIDKey ctxKey = "trace_id"
)

// LoggerKey is method to get logger key
func LoggerKey() any {
	return loggerKey
}

// LogMiddleware is for logging
func LogMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceID := uuid.NewString()

		entry := logrus.WithFields(logrus.Fields{
			"trace_id": traceID,
			"path":     c.Path(),
			"method":   c.Method(),
		})

		c.Locals("logger", entry)

		ctx := context.WithValue(c.UserContext(), loggerKey, entry)
		ctx = context.WithValue(ctx, traceIDKey, traceID)
		c.SetUserContext(ctx)

		return c.Next()
	}
}
