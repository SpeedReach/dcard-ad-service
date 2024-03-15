package cache

import (
	"advertise_service/internal/infra"
	"advertise_service/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"log"
)

func StoreActiveAd(ctx context.Context, ad models.Ad) error {
	rdb := ctx.Value(infra.CacheContextKey{}).(*redis.Client)
	jsonStr := bytes.Buffer{}
	err := json.NewEncoder(&jsonStr).Encode(ad)
	if err != nil {
		return err
	}
	err = rdb.ZAdd(ctx, "active_ads", redis.Z{Member: jsonStr.String(), Score: float64(ad.EndAt.Unix())}).Err()
	if err != nil {
		log.Printf("failed to cache active ad: %v", err)
		return err
	}
	return nil
}
