package cache

import (
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/models"
	"bytes"
	"cmp"
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
	"strconv"
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
	allowStartBefore := time.Now().UTC().Add(Interval).Add(Tolerance)
	ads = slices.DeleteFunc(ads, func(ad models.Ad) bool {
		return !ad.StartAt.Before(allowStartBefore)
	})

	logger := ctx.Value(logging.LoggerContextKey{}).(*zap.Logger)
	//acquire lock to make sure only one client is updating the whole list
	lockId := uuid.New().String()
	err := tryAcquireUpdateLock(ctx, rdb, lockId)
	if err != nil {
		return 0, err
	}
	defer releaseUpdateLock(ctx, rdb, lockId)

	//remove ads that are expired
	removed, err := rdb.ZRemRangeByScore(ctx, adsKey, "-inf", strconv.FormatInt(time.Now().UTC().Unix()/1000, 10)).Result()
	logger.Log(zap.DebugLevel, "removed expired ads", zap.Int64("amount", removed))

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
			logger.Log(zap.DebugLevel, "skipping ad that is already in cache", zap.String("ad_id", ad.ID.String()), zap.Time("start_at", ad.StartAt))
			continue
		} else {
			logger.Log(zap.DebugLevel, "not skipping", zap.Time("start_at", ad.StartAt))
		}
		jsonStr, err := json.Marshal(ad)
		if err != nil {
			return 0, err
		}
		entries = append(entries, redis.Z{Member: jsonStr, Score: float64(ad.EndAt.Unix() / 1000)})
	}

	if len(entries) != 0 {
		//store new ads
		count, err := rdb.ZAdd(ctx, adsKey, entries...).Result()

		if err != nil || count != int64(len(entries)) {
			logger.Log(zap.ErrorLevel, "failed to cache active ads", zap.Error(err), zap.Int64("insert_count", count), zap.String("entries", fmt.Sprint(entries)))
			return 0, err
		}
	}

	//update last update time
	err = rdb.Set(ctx, lastUpdateKey, time.Now().UTC(), time.Hour*2).Err()
	if err != nil {
		logger.Log(zap.ErrorLevel, "fail to update time", zap.Error(err))
		return 0, nil
	}
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
	for _, adStr := range adStrings {
		ad := models.Ad{}
		err := json.NewDecoder(strings.NewReader(adStr)).Decode(&ad)
		if err != nil {
			return ads, err
		}
		ads = append(ads, ad)
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
			return cmp.Compare(i.StartAt.Unix(), j.StartAt.Unix())
		}).StartAt, nil
	} else {
		return time.Unix(0, 0), nil
	}
}
