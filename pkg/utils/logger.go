package utils

import (
	"context"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type loggerKey struct{}

// store the logger in the context for propagation to children
func ContextWithLogger(ctx context.Context, logger logr.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// retrieve the logger from the context
// returns the default logger if not set
func LoggerFromContext(ctx context.Context) logr.Logger {
	logger := ctx.Value(loggerKey{})
	if logger == nil {
		return logf.Log
	}
	log, ok := logger.(logr.Logger)
	if !ok {
		return logf.Log
	}
	return log
}
