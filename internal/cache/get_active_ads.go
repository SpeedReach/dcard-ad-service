package cache

import (
	"advertise_service/internal/infra"
	"advertise_service/internal/models"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"strings"
)

func GetActiveAds(ctx context.Context, skip int, count int) ([]models.Ad, error) {
	rdb := ctx.Value(infra.CacheContextKey{}).(*redis.Client)

	adStrings, err := rdb.ZRange(ctx, "active_ads", int64(skip), int64(skip+count)).Result()
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
