package logger

import (
	"context"
	"os"
	"sync"

	"go.uber.org/zap"
)

var (
	log  *zap.Logger
	once sync.Once
)

// Get initializes (once) and returns a global zap.Logger instance
func Get() *zap.Logger {
	once.Do(func() {
		var err error
		// If in production, use production config
		if env := os.Getenv("ENV"); env == "production" {
			// Use production config to display logs in json format for tracing tools
			log, err = zap.NewProduction()
		} else {
			// Otherwise, use development config
			log, err = zap.NewDevelopment()
		}
		if err != nil {
			panic("Failed to initialize zap logger: " + err.Error())
		}
	})
	return log
}

// --- CONTEXT LOGGING SUPPORT ---

// Unique type used as a context key to prevents key collisions and
// to follow Go’s official best practices for context values.
type contextLoggerKey struct{}

// Inject attaches a logger to the context.
// Use this in HTTP middleware or worker setup.
func Inject(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, contextLoggerKey{}, l)
}

// TODO: Add contexts to all service methods and use FromContext method.
// FromContext retrieves a logger from the context.
// Fallback to the global singleton logger if missing.
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Get()
	}

	if l, ok := ctx.Value(contextLoggerKey{}).(*zap.Logger); ok && l != nil {
		return l
	}

	return Get()
}

// WithFields returns a new context with a logger that includes extra fields.
// Useful for request ID, correlation ID, user ID, etc.
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	l := FromContext(ctx).With(fields...)
	return Inject(ctx, l)
}
