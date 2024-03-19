package cache

import (
	"advertise_service/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

var lastUpdateKey = "lastUpdate"
var adsKey = "active_ads"

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
	err := rdb.Set(ctx, lastUpdateKey, time.Now().UTC(), time.Hour*2).Err()

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
