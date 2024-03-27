package cache

import (
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"log"
	"math"
	"slices"
	"strings"
	"time"
)

func isValid(lastUpdate time.Time) bool {
	return time.Now().UTC().Sub(lastUpdate) < Interval
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

func updateCache(ctx context.Context, rdb *redis.Client, ads []models.Ad) (int, error) {
	logger := ctx.Value(logging.LoggerContextKey{}).(*zap.Logger)
	//acquire lock to make sure only one client is updating the whole list
	lockId := uuid.New().String()
	err := tryAcquireUpdateLock(ctx, rdb, lockId)
	if err != nil {
		return 0, err
	}
	defer releaseUpdateLock(ctx, rdb, lockId)

	//remove ads that are expired
	rdb.ZRemRangeByScore(ctx, adsKey, "-inf", fmt.Sprintf("(%d", time.Now().UTC().Unix()))

	//get the largest start time in the cache
	largestStartTime, err := getLargestStartInCache(ctx, rdb)
	if err != nil {
		return 0, err
	}

	logger.Log(zap.DebugLevel, "largest start time in cache", zap.Time("largest_start_time", largestStartTime))
	//prepare active ads
	var entries []redis.Z
	for _, ad := range ads {
		if !ad.StartAt.After(largestStartTime) {
			logger.Log(zap.DebugLevel, "skipping ad that has already started", zap.String("ad_id", ad.ID.String()), zap.Time("start_at", ad.StartAt))
			continue
		}
		jsonStr, err := json.Marshal(ad)
		if err != nil {
			return 0, err
		}
		entries = append(entries, redis.Z{Member: jsonStr, Score: float64(ad.EndAt.Unix() / 1000)})
	}

	if len(entries) != 0 {
		//store new ads
		err = rdb.ZAdd(ctx, adsKey, entries...).Err()
		if err != nil {
			logger.Log(zap.ErrorLevel, "failed to cache active ads", zap.Error(err), zap.String("entries", fmt.Sprint(entries)))
			return 0, err
		}
	}

	//update last update time
	err = rdb.Set(ctx, lastUpdateKey, time.Now().UTC(), time.Hour*2).Err()
	return len(entries), nil
}

// to make sure only one client updates the whole list, we'll need to implement a simple lock
func tryAcquireUpdateLock(ctx context.Context, client *redis.Client, lockId string) error {
	logger := ctx.Value(logging.LoggerContextKey{}).(*zap.Logger)
	success, err := client.SetNX(ctx, lockKey, lockId, time.Minute).Result()
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
	return client.Del(ctx, lockKey).Err()
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

func getAdsFromRedis(ctx context.Context, client *redis.Client, skip int, count int) ([]models.Ad, error) {
	adStrings, err := client.ZRange(ctx, adsKey, int64(skip), int64(skip+count)).Result()
	if err != nil {
		return []models.Ad{}, err
	}

	ads := make([]models.Ad, 0, len(adStrings))
	now := time.Now().UTC()
	for _, adStr := range adStrings {
		ad := models.Ad{}
		err := json.NewDecoder(strings.NewReader(adStr)).Decode(&ad)
		if err != nil {
			return ads, err
		}

		//cache may contain non-active ads, so we need to filter them out
		if ad.StartAt.Before(now) && ad.EndAt.After(now) {
			ads = append(ads, ad)
		}
	}
	return ads, nil
}

func getLargestStartInCache(ctx context.Context, rdb *redis.Client) (time.Time, error) {
	logger := ctx.Value(logging.LoggerContextKey{}).(*zap.Logger)
	inCacheAds, err := getAdsFromRedis(ctx, rdb, 0, math.MaxInt)
	if err != nil {
		logger.Log(zap.ErrorLevel, "failed to get the largest start time in the cache", zap.Error(err))
		return time.Time{}, err
	}

	if len(inCacheAds) != 0 {
		return slices.MaxFunc(inCacheAds, func(i, j models.Ad) int {
			if i.StartAt.Before(j.StartAt) {
				return -1
			}
			if i.StartAt.After(j.StartAt) {
				return 1
			}
			return 0
		}).StartAt, nil
	} else {
		return time.Unix(0, 0), nil
	}
}
