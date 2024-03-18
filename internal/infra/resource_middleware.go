package infra

import (
	"advertise_service/internal/infra/cache"
	"advertise_service/internal/infra/persistent"
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"net/http"
)

type DatabaseContextKey struct {
}

type CacheContextKey struct {
}

type ResourceMiddleware struct {
	db           persistent.Database
	cacheService cache.Service
}

func (m ResourceMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, DatabaseContextKey{}, m.db)
		ctx = context.WithValue(ctx, CacheContextKey{}, m.cacheService)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewResourceMiddleware(config Config) ResourceMiddleware {
	opt, err := redis.ParseURL(config.RedisURI)
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(opt)

	db, err := sql.Open("pgx", config.PostgresURI)
	if err != nil {
		panic(err)
	}

	storage := persistent.NewSQLDatabase(db)
	return ResourceMiddleware{
		db:           storage,
		cacheService: cache.NewRedisCacheService(redisClient),
	}
}
