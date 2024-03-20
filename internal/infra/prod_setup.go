package infra

import (
	"advertise_service/internal/infra/cache"
	"advertise_service/internal/infra/persistent"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

func ProductionSetup(config Config) (persistent.Storage, cache.Service) {
	opt, err := redis.ParseURL(config.RedisURI)
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(opt)

	db, err := sql.Open("pgx", config.PostgresURI)
	if err != nil {
		panic(err)
	}
	return persistent.NewSQLDatabase(db), cache.NewRedisCacheService(redisClient)
}
