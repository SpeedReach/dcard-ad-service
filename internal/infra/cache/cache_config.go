package cache

import "time"

var (
	lastUpdateKey = "last_update"
	adsKey        = "active_ads"
	lockKey       = "active_ads_lock"

	// Interval is the interval to check if the cache is still valid, we update the cache when it's not valid
	// also we insert ads whose (start time)  < now + (Interval + Tolerance) in to cache
	Interval = time.Hour

	// Tolerance
	// since we use the cache aside method, we may not be able to update the cache successfully right after the cache is invalid,
	// so we need to have a tolerance for the cache to be invalid
	Tolerance = 20 * time.Minute
)
