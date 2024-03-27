package persistent

import (
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/models"
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (db database) InsertAd(ctx context.Context, ad models.Ad) error {
	logger := ctx.Value(logging.LoggerContextKey{}).(*zap.Logger)
	_, err := db.inner.ExecContext(ctx, "INSERT INTO Ads (id, title, start_at, end_at) VALUES ($1, $2, $3, $4)", ad.ID, ad.Title, ad.StartAt, ad.EndAt)
	if err != nil {
		logger.Log(zap.ErrorLevel, "Could not execute context for insert ad", zap.Error(err))
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

func (db database) insertCondition(ctx context.Context, parentAdID uuid.UUID, condition models.Condition) error {
	logger := ctx.Value(logging.LoggerContextKey{}).(*zap.Logger)
	stmt, err := db.inner.PrepareContext(ctx, "INSERT INTO Conditions (id, ad_id, ios, android, web, jp, tw, male, female, min_age, max_age) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)")
	if err != nil {
		logger.Log(zap.ErrorLevel, "Could not prepare context for insert condition", zap.Error(err))
		return err
	}
	defer stmt.Close()
	schema := FromConditionModel(condition)

	_, err = stmt.ExecContext(ctx, uuid.New(), parentAdID,
		schema.Ios, schema.Android, schema.Web, schema.Jp, schema.Tw, schema.Male, schema.Female, schema.MinAge, schema.MaxAge)

	if err != nil {
		logger.Log(zap.ErrorLevel, "Could not insert condition", zap.Error(err))
		return err
	}
	return nil
}
