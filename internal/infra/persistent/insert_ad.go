package persistent

import (
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/models"
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (db Database) InsertAd(ctx context.Context, ad models.Ad) error {
	stmt, err := db.inner.PrepareContext(ctx, "INSERT INTO Ads (id, title, start_at, end_at) VALUES ($1, $2, $3, $4)")
	if err != nil {
		logging.ContextualLog(ctx, zap.ErrorLevel, "Could not prepare context for insert ad", zap.Error(err))
		return err
	}

	_, err = stmt.ExecContext(ctx, ad.ID, ad.Title, ad.StartAt, ad.EndAt)
	if err != nil {
		logging.ContextualLog(ctx, zap.ErrorLevel, "Could not execute context for insert ad", zap.Error(err))
		return err
	}

	for _, condition := range ad.Conditions {
		err = db.insertCondition(ctx, ad.ID, condition)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db Database) insertCondition(ctx context.Context, parentAdID uuid.UUID, condition models.Condition) error {
	stmt, err := db.inner.PrepareContext(ctx, "INSERT INTO Conditions (id, ad_id, ios, android, web, jp, tw, male, female, min_age, max_age) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)")
	if err != nil {
		logging.ContextualLog(ctx, zap.ErrorLevel, "Could not prepare context for insert condition", zap.Error(err))
		return err
	}
	schema := FromConditionModel(condition)

	_, err = stmt.ExecContext(ctx, uuid.New(), parentAdID,
		schema.Ios, schema.Android, schema.Web, schema.Jp, schema.Tw, schema.Male, schema.Female, schema.MinAge, schema.MaxAge)

	if err != nil {
		logging.ContextualLog(ctx, zap.ErrorLevel, "Could not insert condition", zap.Error(err))
		return err
	}
	return nil
}
