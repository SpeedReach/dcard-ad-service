//go:build test

package mock

import (
	"advertise_service/internal/infra/cache"
	"advertise_service/internal/models"
	"cmp"
	"context"
	"slices"
	"time"
)

type mockCache struct {
	inner *cacheArray
}

func (c mockCache) Clear(ctx context.Context) error {
	c.inner.ads = []models.Ad{}
	return nil
}

type cacheArray struct {
	ads []models.Ad
}

func NewCache() cache.Service {
	return mockCache{
		inner: &cacheArray{},
	}
}

func (c mockCache) CheckCacheValid(ctx context.Context) (bool, error) {
	return len(c.inner.ads) > 0, nil
}

// GetActiveAds retrieves active ads with params skip and count in a sorted list.
func (c mockCache) GetActiveAds(ctx context.Context, skip int, count int) ([]models.Ad, error) {
	l := len(c.inner.ads)
	return c.inner.ads[min(l, skip):min(l, skip+count)], nil
}

// WriteActiveAd stores an active ad into the mockCache
func (c mockCache) WriteActiveAd(ctx context.Context, ad models.Ad) error {
	c.inner.ads = append(c.inner.ads, ad)
	slices.SortFunc(c.inner.ads, func(a, b models.Ad) int {
		if a.EndAt.After(b.EndAt) {
			return 1
		} else {
			return -1
		}
	})
	return nil
}

// Update stores multiple active ads into mockCache
func (c mockCache) Update(ctx context.Context, ad []models.Ad) (int, error) {
	c.inner.ads = slices.DeleteFunc(c.inner.ads, func(a models.Ad) bool {
		return a.EndAt.Before(time.Now().UTC())
	})

	var largestStart time.Time = time.Unix(0, 0).UTC()

	if len(c.inner.ads) != 0 {
		largestStart = slices.MaxFunc(c.inner.ads, func(a, b models.Ad) int {
			return cmp.Compare(a.StartAt.Unix(), b.StartAt.Unix())
		}).StartAt
	}

	ad = slices.DeleteFunc(ad, func(a models.Ad) bool {
		return !a.StartAt.After(largestStart)
	})

	c.inner.ads = append(c.inner.ads, ad...)
	slices.SortFunc(c.inner.ads, func(a, b models.Ad) int {
		if a.EndAt.After(b.EndAt) {
			return 1
		} else {
			return -1
		}
	})

	return len(ad), nil
}
