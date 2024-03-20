package persistent

import (
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/models"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
	"time"
)

type Storage interface {
	InsertAd(ctx context.Context, ad models.Ad) error
	FindAdsWithTime(ctx context.Context, startBefore time.Time, endAfter time.Time) ([]models.Ad, error)
}

func TestStorage(t *testing.T, db Storage) {
	logger, _ := zap.NewDevelopment()
	ctx := context.WithValue(context.Background(), logging.LoggerContextKey{}, logger)
	now := time.Now().UTC()
	ad := models.Ad{
		ID:      uuid.New(),
		Title:   "test",
		StartAt: now.Add(-time.Hour),
		EndAt:   now.Add(time.Hour),
		Conditions: []models.Condition{
			{
				AgeStart: 20,
			},
		},
	}

	ad2 := models.Ad{
		ID:      uuid.New(),
		Title:   "test2",
		StartAt: now.Add(-2 * time.Hour),
		EndAt:   now.Add(-time.Hour),
	}

	t.Run("InsertAd", func(t *testing.T) {
		err := db.InsertAd(ctx, ad)
		require.NoError(t, err)
		err = db.InsertAd(ctx, ad2)
		require.NoError(t, err)
	})

	t.Run("FindAdsWithTime", func(t *testing.T) {
		ads, err := db.FindAdsWithTime(ctx, now, now)
		require.NoError(t, err)
		require.Len(t, ads, 1)
		require.Equal(t, ad.Title, ads[0].Title)
	})

}
