//go:build test

package handlers

import (
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/mock"
	"context"
	"go.uber.org/zap"
)

var (
	staticStorage = mock.NewStorage()
	staticCache   = mock.NewCache()
	logger, _     = zap.NewDevelopment()
)

func InjectStaticMockedResources(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, StorageContextKey{}, staticStorage)
	ctx = context.WithValue(ctx, CacheContextKey{}, staticCache)
	ctx = context.WithValue(ctx, logging.LoggerContextKey{}, logger)
	return ctx
}

func InjectMockedResources(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, StorageContextKey{}, mock.NewStorage())
	ctx = context.WithValue(ctx, CacheContextKey{}, mock.NewCache())
	ctx = context.WithValue(ctx, logging.LoggerContextKey{}, logger)
	return ctx
}
