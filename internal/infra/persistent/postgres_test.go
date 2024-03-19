//go:build integration

package persistent

import (
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/models"
	"context"
	"database/sql"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"os"
	"slices"
	"testing"
	"time"
)

func TestPostgresInsertAndGet(t *testing.T) {
	logger, _ := zap.NewProduction()
	ctx := context.WithValue(context.Background(), logging.LoggerContextKey{}, logger)
	pgUri, found := os.LookupEnv("POSTGRES_URI")
	if !found {
		t.Error("POSTGRES_URI not found")
		return
	}
	sqlDb, err := sql.Open("pgx", pgUri)
	assert.NoError(t, err)
	db := NewSQLDatabase(sqlDb)

	now := time.Now().UTC()
	ad1 := models.Ad{
		ID:      uuid.New(),
		Title:   "title1",
		StartAt: now.Add(-2 * time.Hour),
		EndAt:   now.Add(time.Hour),
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
	ad2 := models.Ad{
		ID:      uuid.New(),
		Title:   "title2",
		StartAt: now.Add(-3 * time.Hour),
		EndAt:   now.Add(-2 * time.Hour),
	}

	err = db.InsertAd(context.Background(), ad1)
	assert.NoError(t, err)
	err = db.InsertAd(context.Background(), ad2)
	assert.NoError(t, err)

	ads, err := db.FindAdsWithTime(ctx, now.Add(-time.Hour), now)
	assert.NoError(t, err)
	ad1Index := slices.IndexFunc(ads, func(i models.Ad) bool {
		return i.Title == "title1"
	})
	assert.NotEqual(t, -1, ad1Index)
	ad2Index := slices.IndexFunc(ads, func(i models.Ad) bool {
		return i.Title == "title2"
	})
	assert.Equal(t, -1, ad2Index)
	assert.Equal(t, ad1.Conditions, ads[ad1Index].Conditions, ad1.ID)

}
