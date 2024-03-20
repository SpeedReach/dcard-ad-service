package persistent

import (
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/models"
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"slices"
	"time"
)

func (db database) FindAdsWithTime(ctx context.Context, startBefore time.Time, endAfter time.Time) ([]models.Ad, error) {
	logger := ctx.Value(logging.LoggerContextKey{}).(*zap.Logger)
	prepareContext, err := db.inner.PrepareContext(ctx, `
			SELECT a.id, a.title, a.start_at, a.end_at, c.min_age, c.max_age, c.male, c.female, c.ios, c.android, c.web, c.jp, c.tw
			FROM Ads a
			LEFT JOIN Conditions c ON a.id = c.ad_id
			WHERE a.start_at < $1 AND a.end_at > $2
		`)

	if err != nil {
		logger.Log(zap.ErrorLevel, "Could not prepare context for find ads with time", zap.Error(err))
		return []models.Ad{}, err
	}
	rows, err := prepareContext.QueryContext(ctx, startBefore, endAfter)
	if err != nil {
		logger.Log(zap.ErrorLevel, "Could not query context for find ads with time", zap.Error(err))
		return []models.Ad{}, err
	}
	defer rows.Close()

	ads := map[uuid.UUID]models.Ad{}
	for rows.Next() {
		ad := models.Ad{}
		condition := ScannedCondition{}
		err = rows.Scan(&ad.ID, &ad.Title, &ad.StartAt, &ad.EndAt,
			&condition.MinAge, &condition.MaxAge, &condition.Male, &condition.Female, &condition.Ios, &condition.Android, &condition.Web, &condition.Jp, &condition.Tw)
		if err != nil {
			logger.Log(zap.ErrorLevel, "error scanning rows", zap.Error(err))
			return []models.Ad{}, err
		}
		if _, ok := ads[ad.ID]; !ok {
			ad.Conditions = []models.Condition{ToConditionModel(condition)}
			ads[ad.ID] = ad
		} else {
			ad.Conditions = append(ads[ad.ID].Conditions, ToConditionModel(condition))
			ads[ad.ID] = ad
		}

	}

	if rows.Err() != nil {
		logger.Log(zap.ErrorLevel, "error scanning rows", zap.Error(err))
		return []models.Ad{}, err
	}
	values := make([]models.Ad, len(ads))

	i := 0
	for _, v := range ads {
		values[i] = v
		i++
	}
	slices.SortFunc(values, func(i, j models.Ad) int {
		if i.EndAt.Before(j.EndAt) {
			return -1
		}
		return 1
	})
	return values, nil
}
