package cache

import (
	"advertise_service/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log"
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
	//make sure only one client is updating the whole list
	lockId := uuid.New().String()
	err := tryAcquireUpdateLock(ctx, rdb, lockId)
	if err != nil {
		return err
	}
	err = rdb.Set(ctx, lastUpdateKey, time.Now().UTC(), time.Hour*2).Err()
	_ = releaseUpdateLock(ctx, rdb, lockId)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	entries := make([]redis.Z, len(ads))
	for i, ad := range ads {
		jsonStr, err := json.Marshal(ad)
		if err != nil {
			return err
		}
		entries[i] = redis.Z{Member: jsonStr, Score: float64(ad.EndAt.Unix() / 1000)}
	}
	rdb.ZAdd(ctx, adsKey, entries...)
	return nil
}

// to make sure only one client updates the whole list, we'll need to implement a simple lock
func tryAcquireUpdateLock(ctx context.Context, client *redis.Client, lockId string) error {
	success, err := client.SetNX(ctx, "active_ads_lock", lockId, 500*time.Millisecond).Result()
	if err != nil {
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
