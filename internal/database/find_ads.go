package database

import (
	"advertise_service/internal/infra"
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/models"
	"context"
	"database/sql"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func FindAdsWithTime(ctx context.Context) ([]models.Ad, error) {
	db := ctx.Value(infra.DatabaseContextKey{}).(*sql.DB)

	prepareContext, err := db.PrepareContext(ctx, `
			SELECT a.id, a.title, a.start_at, a.end_at, c.min_age, c.max_age, c.male, c.female, c.ios, c.android, c.web
			FROM Ads a
			LEFT JOIN Conditions c ON a.id = c.ad_id
			WHERE a.start_at < NOW() AND a.end_at > NOW()		
		`)
	if err != nil {
		logging.ContextualLog(ctx, zap.ErrorLevel, "Could not prepare context for find ads with time", zap.Error(err))
		return []models.Ad{}, err
	}
	rows, err := prepareContext.QueryContext(ctx)
	if err != nil {
		logging.ContextualLog(ctx, zap.ErrorLevel, "Could not query context for find ads with time", zap.Error(err))
		return []models.Ad{}, err
	}
	defer rows.Close()

	ads := map[uuid.UUID]models.Ad{}
	for rows.Next() {
		ad := models.Ad{}
		condition := ConditionSchema{}
		err = rows.Scan(&ad.ID, &ad.Title, &ad.StartAt, &ad.EndAt,
			&condition.MinAge, &condition.MaxAge, &condition.Male, &condition.Female, &condition.Ios, &condition.Android, &condition.Web)
		if err != nil {
			logging.ContextualLog(ctx, zap.ErrorLevel, "error scanning rows", zap.Error(err))
			return []models.Ad{}, err
		}
		if _, ok := ads[ad.ID]; !ok {
			ads[ad.ID] = ad
		} else {
			ad.Conditions = append(ads[ad.ID].Conditions, ToConditionModel(condition))
			ads[ad.ID] = ad
		}
	}
	values := make([]models.Ad, len(ads))
	i := 0
	for _, v := range ads {
		values[i] = v
		i++
	}

	return values, nil
}
