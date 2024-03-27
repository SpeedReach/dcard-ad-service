//go:build test

package mock

import (
	"advertise_service/internal/infra/persistent"
	"advertise_service/internal/models"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type storage struct {
	inner *innerStorage
}

type innerStorage struct {
	ads map[uuid.UUID]models.Ad
}

func NewStorage() persistent.Storage {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	persistent.CreateTables(db)
	if err != nil {
		return nil
	}

	return persistent.NewSQLDatabase(db)
}

func (s storage) InsertAd(ctx context.Context, ad models.Ad) error {
	if _, ok := s.inner.ads[ad.ID]; ok {
		return errors.New("ad already exists")
	}
	s.inner.ads[ad.ID] = ad
	return nil
}

func (s storage) FindAdsWithTime(ctx context.Context, startBefore time.Time, endAfter time.Time) ([]models.Ad, error) {
	var results []models.Ad
	for _, ad := range s.inner.ads {
		if ad.StartAt.Before(startBefore) && ad.EndAt.After(endAfter) {
			results = append(results, ad)
		}
	}
	return results, nil
}
