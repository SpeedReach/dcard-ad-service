package logging

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"time"
)

type LoggerMiddleware struct {
	Logger *zap.Logger
}

type RequestIdContextKey struct{}
type LoggerContextKey struct{}

func (m LoggerMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := uuid.New().String()
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestIdContextKey{}, requestId)
		ctx = context.WithValue(ctx, LoggerContextKey{}, m.Logger)
		r.WithContext(ctx)

		fields := []zap.Field{
			zap.Time("time", time.Now()),
			zap.String("userAgent", r.UserAgent()),
			zap.String("method", r.Method),
			zap.String("uri", r.URL.Path),
			zap.String("requestId", requestId),
			zap.Int64("bytesIn", r.ContentLength),
		}
		start := time.Now()
		next.ServeHTTP(w, r.WithContext(ctx))

		fields = append(fields,
			zap.Duration("latency", time.Since(start)),
		)
		m.Logger.Info("Request", fields...)
		return
	})
}

func ContextualLog(ctx context.Context, lvl zapcore.Level, msg string, fields ...zap.Field) {
	logger := ctx.Value(LoggerContextKey{}).(*zap.Logger)
	if ctx.Value(RequestIdContextKey{}) == nil {
		logger.Log(lvl, msg, fields...)
	} else {
		requestId := ctx.Value(RequestIdContextKey{}).(string)
		logger.With(zap.String("requestId", requestId)).Log(lvl, msg, fields...)
	}
}
