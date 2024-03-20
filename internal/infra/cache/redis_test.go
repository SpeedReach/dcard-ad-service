//go:build integration

package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
	"testing"
	"time"
)

func setup(t *testing.T) *redis.Client {
	redisURI, found := os.LookupEnv("REDIS_URI")
	if !found {
		t.Error("REDIS_URI not found")
		return nil
	}
	opt, err := redis.ParseURL(redisURI)
	if err != nil {
		panic(err)
	}
	return redis.NewClient(opt)
}

func reset(rdb *redis.Client) {
	rdb.Del(context.Background(), lastUpdateKey)
	rdb.Del(context.Background(), adsKey)
}

func TestRedis(t *testing.T) {
	rdb := setup(t)

	defer rdb.Close()

	reset(rdb)
	rdb.Set(context.Background(), lastUpdateKey, time.Unix(0, 0), time.Hour*2)

	service := NewRedisCacheService(rdb)
	TestCacheService(t, service)
}
