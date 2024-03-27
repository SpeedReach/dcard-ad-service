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
	"slices"
	"testing"
	"time"
)

type Service interface {
	// CheckCacheValid checks if the cache is updated within an hour
	CheckCacheValid(ctx context.Context) (bool, error)
	// GetActiveAds retrieves active ads with params skip and count in a sorted list.
	GetActiveAds(ctx context.Context, skip int, count int) ([]models.Ad, error)

	// WriteActiveAd stores an active ad into the cache, used when the create ad is already active.
	WriteActiveAd(ctx context.Context, ad models.Ad) error

	// Update updates lastUpdate time, clears expired ads, and inserts new multiple active ads into cache.
	// NOTICE: This function only writes ad that has a start time greater than the largest start time in the cache
	Update(ctx context.Context, ad []models.Ad) (int, error)

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
	t, err := getLastUpdate(ctx, r.inner)
	if err != nil {
		return false, err
	}

	return isValid(t), nil
}

func (r redisCacheService) GetActiveAds(ctx context.Context, skip int, count int) ([]models.Ad, error) {
	ads, err := getAdsFromRedis(ctx, r.inner, skip, count)
	if err != nil {
		return []models.Ad{}, err
	}
	//may contain ads that start later, so we need to filter them out.
	now := time.Now()
	ads = slices.DeleteFunc(ads, func(ad models.Ad) bool {
		return ad.StartAt.After(now)
	})
	return ads, nil
}

func (r redisCacheService) WriteActiveAd(ctx context.Context, ad models.Ad) error {
	return storeActiveAd(ctx, r.inner, ad)
}

func (r redisCacheService) Clear(ctx context.Context) error {
	return r.inner.Del(ctx, lastUpdateKey, adsKey).Err()
}

func (r redisCacheService) Update(ctx context.Context, ads []models.Ad) (int, error) {
	return updateCache(ctx, r.inner, ads)
}

func TestCacheService(t *testing.T, service Service) {
	logger, _ := zap.NewDevelopment()
	ctx := context.WithValue(context.Background(), logging.LoggerContextKey{}, logger)
	require.NoError(t, service.Clear(ctx))

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
		err := service.WriteActiveAd(ctx, ad)
		require.NoError(t, err)

		activeAds, err := service.GetActiveAds(ctx, 0, 3)
		if err != nil {
			return
		}
		require.Len(t, activeAds, 1)
		assert.Equal(t, ad.ID, activeAds[0].ID)
		assert.Equal(t, ad.Conditions, activeAds[0].Conditions)
	})
	require.NoError(t, service.Clear(ctx))

	t.Run("Update", func(t *testing.T) {
		assert.NoError(t, err)
		//write many
		ads := []models.Ad{
			{
				ID:      uuid.New(),
				Title:   "title1",
				StartAt: time.Now().UTC().Add(-2 * Interval),
				EndAt:   time.Now().UTC().Add(2 * Interval),
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
				ID:      uuid.New(),
				StartAt: time.Now().UTC().Add(-2 * Interval),
				Title:   "title2",
				EndAt:   time.Now().UTC().Add(1 * Interval),
			},
		}

		writeCount, err := service.Update(ctx, ads)
		require.NoError(t, err)
		assert.Equal(t, 2, writeCount)

		valid, err = service.CheckCacheValid(ctx)
		require.NoError(t, err)
		assert.True(t, valid)

		activeAds, err := service.GetActiveAds(ctx, 0, 3)
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

		//write second time
		writeCount, err = service.Update(ctx, ads)
		assert.Equal(t, 0, writeCount)
	})

}
