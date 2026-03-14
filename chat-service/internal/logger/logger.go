package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
}

type contextKey string

const RequestIDKey contextKey = "request_id"

type Config struct {
	ServiceName string
	Environment string
	PrettyPrint bool
	Output      io.Writer
}

func New(cfg Config) *Logger {
	var output io.Writer = os.Stdout
	if cfg.Output != nil {
		output = cfg.Output
	}

	if cfg.PrettyPrint {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}
	}

	logger := zerolog.New(output).
		With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Str("environment", cfg.Environment).
		Logger()

	return &Logger{logger}
}

func (log *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{log.With().Str("request_id", requestID).Logger()}
}

func (log *Logger) FromContext(ctx context.Context) *Logger {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return log.WithRequestID(requestID)
	}
	return log
}

func GetRequestID(ctx context.Context) string {
	if val, ok := ctx.Value(RequestIDKey).(string); ok {
		return val
	}
	return ""
}

func SetRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return context.WithValue(ctx, RequestIDKey, requestID)
}
