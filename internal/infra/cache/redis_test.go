//go:build integration

package cache

import (
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/models"
	"context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"os"
	"slices"
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
	rdb.Del(context.Background(), adsKey, lastUpdateKey, lockKey)
}

func TestRedis(t *testing.T) {
	testData := []models.Ad{
		{
			ID:      uuid.New(),
			StartAt: time.Now().UTC().Add(-Interval),
			EndAt:   time.Now().UTC().Add(Interval),
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().UTC().Add(max(time.Minute, Tolerance)),
			EndAt:   time.Now().UTC().Add(Interval),
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().UTC().Add(max(time.Minute, Interval+2*Tolerance)),
			EndAt:   time.Now().UTC().Add(Interval),
		},
	}

	t.Run("lock test", func(t *testing.T) {
		rdb := setup(t)
		reset(rdb)
		defer rdb.Close()
		logger, _ := zap.NewDevelopment()
		done := make(chan bool)
		ctx := context.WithValue(context.Background(), logging.LoggerContextKey{}, logger)
		go func() {
			lockId := uuid.New().String()
			assert.NoError(t, tryAcquireUpdateLock(ctx, rdb, lockId))
			time.Sleep(time.Millisecond * 500)
			assert.NoError(t, releaseUpdateLock(ctx, rdb, lockId))
		}()

		go func() {
			time.Sleep(time.Millisecond * 10)
			assert.Error(t, tryAcquireUpdateLock(ctx, rdb, uuid.New().String()))
			time.Sleep(time.Second * 1)
			lockId := uuid.New().String()
			assert.NoError(t, tryAcquireUpdateLock(ctx, rdb, lockId))
			assert.NoError(t, releaseUpdateLock(ctx, rdb, lockId))
			done <- true
		}()
		<-done
	})

	t.Run("update & get", func(t *testing.T) {
		rdb := setup(t)
		reset(rdb)
		defer rdb.Close()
		logger, _ := zap.NewDevelopment()
		ctx := context.WithValue(context.Background(), logging.LoggerContextKey{}, logger)
		writeAmount, err := updateCache(ctx, rdb, slices.Clone(testData))
		require.NoError(t, err)
		assert.Equal(t, 2, writeAmount)

		ads, err := getAdsFromRedis(ctx, rdb, 0, 1000)
		require.NoError(t, err)
		assert.Equal(t, 2, len(ads))
	})

	t.Run("second update", func(t *testing.T) {
		rdb := setup(t)
		reset(rdb)
		defer rdb.Close()
		logger, _ := zap.NewDevelopment()
		ctx := context.WithValue(context.Background(), logging.LoggerContextKey{}, logger)

		writeAmount, err := updateCache(ctx, rdb, slices.Clone(testData))
		require.NoError(t, err)
		assert.Equal(t, 2, writeAmount)

		writeAmount, err = updateCache(ctx, rdb, testData)
		require.NoError(t, err)
		assert.Equal(t, 0, writeAmount)
	})

	t.Run("full test", func(t *testing.T) {
		rdb := setup(t)
		reset(rdb)
		defer rdb.Close()
		TestCacheService(t, NewRedisCacheService(rdb))
	})
}
