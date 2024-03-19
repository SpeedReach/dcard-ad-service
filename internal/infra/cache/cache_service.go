package cache

import (
	"advertise_service/internal/models"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type Service interface {
	// CheckCacheValid checks if the cache is updated within an hour
	CheckCacheValid(ctx context.Context) (bool, error)
	// GetActiveAds retrieves active ads with params skip and count
	GetActiveAds(ctx context.Context, skip int, count int) ([]models.Ad, error)
	// WriteActiveAd stores an active ad into the cache
	WriteActiveAd(ctx context.Context, ad models.Ad) error
	WriteActiveAds(ctx context.Context, ad []models.Ad) error
}

type redisCacheService struct {
	inner *redis.Client
}

func NewRedisCacheService(inner *redis.Client) Service {
	inner.Del(context.Background(), lastUpdateKey)
	inner.Del(context.Background(), adsKey)
	return redisCacheService{inner: inner}
}

func (r redisCacheService) CheckCacheValid(ctx context.Context) (bool, error) {
	result, err := r.inner.Get(ctx, lastUpdateKey).Result()

	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	t, err := time.Parse(time.RFC3339Nano, result)
	if err != nil {
		return false, nil
	}

	return time.Now().UTC().Sub(t) < 1*time.Hour, nil
}

func (r redisCacheService) GetActiveAds(ctx context.Context, skip int, count int) ([]models.Ad, error) {
	adStrings, err := r.inner.ZRange(ctx, adsKey, int64(skip), int64(skip+count)).Result()
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

func (r redisCacheService) WriteActiveAd(ctx context.Context, ad models.Ad) error {
	return storeActiveAd(ctx, r.inner, ad)
}

func (r redisCacheService) WriteActiveAds(ctx context.Context, ads []models.Ad) error {
	return storeActiveAds(ctx, r.inner, ads)
}
