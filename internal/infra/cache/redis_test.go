//go:build integration

package cache

import (
	"advertise_service/internal/models"
	"context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
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

	valid, err := service.CheckCacheValid(context.Background())
	assert.NoError(t, err)
	assert.False(t, valid)

	//write many
	ads := []models.Ad{
		{
			ID:    uuid.New(),
			Title: "title1",
			EndAt: time.Now().Add(2 * time.Hour),
		},
		{
			ID:    uuid.New(),
			Title: "title2",
			EndAt: time.Now().Add(1 * time.Hour),
		},
	}

	err = service.WriteActiveAds(context.Background(), ads)
	assert.NoError(t, err)

	valid, err = service.CheckCacheValid(context.Background())
	assert.NoError(t, err)
	assert.True(t, valid)

	activeAds, err := service.GetActiveAds(context.Background(), 0, 3)
	if err != nil {
		return
	}
	assert.Len(t, activeAds, 2)
	//check if sorted
	assert.Equal(t, ads[0].ID, activeAds[1].ID)
	assert.Equal(t, ads[1].ID, activeAds[0].ID)
}
