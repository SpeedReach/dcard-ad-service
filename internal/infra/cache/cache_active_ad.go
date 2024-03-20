package cache

import (
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"log"
	"slices"
	"strconv"
	"time"
)

var lastUpdateKey = "lastUpdate"
var adsKey = "active_ads"

func isValid(lastUpdate time.Time) bool {
	return time.Now().UTC().Sub(lastUpdate) < 1*time.Hour
}

func storeActiveAd(ctx context.Context, rdb *redis.Client, ad models.Ad) error {
	jsonStr := bytes.Buffer{}
	err := json.NewEncoder(&jsonStr).Encode(ad)
	if err != nil {
		return err
	}
	err = rdb.ZAdd(ctx, adsKey, redis.Z{Member: jsonStr.String(), Score: float64(ad.EndAt.Unix())}).Err()
	if err != nil {
		log.Printf("failed to cache active ad: %v", err)
		return err
	}
	return nil
}

func storeActiveAds(ctx context.Context, rdb *redis.Client, ads []models.Ad) error {
	logger := ctx.Value(logging.LoggerContextKey{}).(*zap.Logger)
	lastUpdate, err := getLastUpdate(ctx, rdb)

	//make sure only one client is updating the whole list
	lockId := uuid.New().String()
	err = tryAcquireUpdateLock(ctx, rdb, lockId)
	if err != nil {
		return err
	}

	//update last update time
	err = rdb.Set(ctx, lastUpdateKey, time.Now().UTC(), time.Hour*2).Err()
	_ = releaseUpdateLock(ctx, rdb, lockId)
	if err != nil {
		logger.Log(zap.ErrorLevel, "failed to update last update time", zap.Error(err))
		return err
	}

	//clear expired ads
	err = rdb.ZRemRangeByScore(ctx, adsKey, "0", strconv.Itoa(int(lastUpdate.Unix()/1000))).Err()
	if err != nil {
		logger.Log(zap.ErrorLevel, "failed to clear expired ads", zap.Error(err))
		return err
	}

	//remove duplicate ads
	ads = slices.DeleteFunc(ads, func(i models.Ad) bool {
		return lastUpdate.Add(80 * time.Minute).Before(i.StartAt)
	})
	entries := make([]redis.Z, len(ads))
	for i, ad := range ads {
		jsonStr, err := json.Marshal(ad)
		if err != nil {
			return err
		}
		entries[i] = redis.Z{Member: jsonStr, Score: float64(ad.EndAt.Unix() / 1000)}
	}
	err = rdb.ZAdd(ctx, adsKey, entries...).Err()
	if err != nil {
		logger.Log(zap.ErrorLevel, "failed to cache active ads", zap.Error(err))
		return err
	}
	return nil
}

// to make sure only one client updates the whole list, we'll need to implement a simple lock
func tryAcquireUpdateLock(ctx context.Context, client *redis.Client, lockId string) error {
	logger := ctx.Value(logging.LoggerContextKey{}).(*zap.Logger)
	success, err := client.SetNX(ctx, "active_ads_lock", lockId, 500*time.Millisecond).Result()
	if err != nil {
		logger.Log(zap.ErrorLevel, "failed to acquire lock", zap.Error(err))
		return err
	}
	if !success {
		return errors.New("lock is already acquired by someone else")
	}

	return nil
}

func releaseUpdateLock(ctx context.Context, client *redis.Client, lockId string) error {
	lockIdInCache, err := client.Get(ctx, "active_ads_lock").Result()
	if err != nil {
		return err
	}

	if lockIdInCache != lockId {
		return errors.New("lock id doesn't match, this is not your lock")
	}
	return client.Del(ctx, "active_ads_lock").Err()
}

func getLastUpdate(ctx context.Context, client *redis.Client) (time.Time, error) {
	result, err := client.Get(ctx, lastUpdateKey).Result()

	if err != nil {
		if err == redis.Nil {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339Nano, result)
}
