package cache

import (
	"advertise_service/internal/models"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

type Service interface {
	// CheckCacheValid checks if the cache is updated within an hour
	CheckCacheValid(ctx context.Context) (bool, error)
	// GetActiveAds retrieves active ads with params skip and count in a sorted list.
	GetActiveAds(ctx context.Context, skip int, count int) ([]models.Ad, error)
	// WriteActiveAd stores an active ad into the cache
	WriteActiveAd(ctx context.Context, ad models.Ad) error
	// WriteActiveAds stores multiple active ads into cache
	WriteActiveAds(ctx context.Context, ad []models.Ad) error
	// Clear clears the cache, useful for testing
	Clear(ctx context.Context) error
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

	return isValid(t), nil
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

func (r redisCacheService) Clear(ctx context.Context) error {
	return r.inner.Del(ctx, lastUpdateKey, adsKey).Err()
}

func (r redisCacheService) WriteActiveAds(ctx context.Context, ads []models.Ad) error {
	return storeActiveAds(ctx, r.inner, ads)
}

func TestCacheService(t *testing.T, service Service) {
	err := service.Clear(context.Background())
	require.NoError(t, err)
	valid, err := service.CheckCacheValid(context.Background())
	require.NoError(t, err)
	assert.False(t, valid)

	t.Run("WriteActiveAd", func(t *testing.T) {
		ad := models.Ad{
			ID:      uuid.New(),
			Title:   "title1",
			StartAt: time.Now().UTC().Add(-2 * time.Hour),
			EndAt:   time.Now().UTC().Add(time.Hour),
			Conditions: []models.Condition{
				{
					AgeStart: 20,
					AgeEnd:   30,
					Country: []models.Country{
						models.Taiwan,
					},
				},
			},
		}
		err := service.WriteActiveAd(context.Background(), ad)
		require.NoError(t, err)

		activeAds, err := service.GetActiveAds(context.Background(), 0, 3)
		if err != nil {
			return
		}
		require.Len(t, activeAds, 1)
		assert.Equal(t, ad.ID, activeAds[0].ID)
		assert.Equal(t, ad.Conditions, activeAds[0].Conditions)
	})

	t.Run("WriteActiveAds", func(t *testing.T) {
		err := service.Clear(context.Background())
		assert.NoError(t, err)
		//write many
		ads := []models.Ad{
			{
				ID:    uuid.New(),
				Title: "title1",
				EndAt: time.Now().UTC().Add(2 * time.Hour),
				Conditions: []models.Condition{
					{
						AgeStart: 20,
						AgeEnd:   30,
						Country: []models.Country{
							models.Taiwan,
						},
						Platform: []models.Platform{
							models.Web,
						},
						Gender: []models.Gender{
							models.Male,
						},
					},
				},
			},
			{
				ID:    uuid.New(),
				Title: "title2",
				EndAt: time.Now().UTC().Add(1 * time.Hour),
			},
		}

		err = service.WriteActiveAds(context.Background(), ads)
		require.NoError(t, err)

		valid, err = service.CheckCacheValid(context.Background())
		assert.NoError(t, err)
		assert.True(t, valid)

		activeAds, err := service.GetActiveAds(context.Background(), 0, 3)
		if err != nil {
			return
		}
		require.Len(t, activeAds, 2)
		//check if sorted
		assert.Equal(t, ads[0].ID, activeAds[1].ID)
		assert.Equal(t, ads[1].ID, activeAds[0].ID)
		assert.Equal(t, activeAds[1].Conditions[0].Country[0], models.Taiwan)
		assert.Equal(t, activeAds[1].Conditions[0].Platform[0], models.Web)
		assert.Equal(t, activeAds[1].Conditions[0].Gender[0], models.Male)
		assert.Equal(t, activeAds[1].Conditions[0].AgeStart, 20)
		assert.Equal(t, activeAds[1].Conditions[0].AgeEnd, 30)
	})

}
