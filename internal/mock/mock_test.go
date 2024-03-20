package mock

import (
	"advertise_service/internal/infra/cache"
	"advertise_service/internal/infra/persistent"
	"testing"
)

func TestMockStorage(t *testing.T) {
	persistent.TestStorage(t, NewStorage())
}

func TestMockCache(t *testing.T) {
	cache.TestCacheService(t, NewCache())
}
