package handlers

import (
	"advertise_service/internal/infra/persistent"
	"advertise_service/internal/mock"
	"advertise_service/internal/models"
	"advertise_service/internal/utils"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestPostAd(t *testing.T) {
	testData := mock.GenerateMockAds()
	testData = slices.DeleteFunc(testData, func(ad models.Ad) bool {
		return strings.Contains(ad.Title, "inactive")
	})
	utils.SortAdsByID(testData)
	ctx := InjectMockedResources(context.Background())

	for _, ad := range testData {
		request := PostAdRequest{
			Title:      ad.Title,
			StartAt:    ad.StartAt,
			EndAt:      ad.EndAt,
			Conditions: ad.Conditions,
		}
		_, err := postAd(ctx, request)
		require.NoError(t, err)
	}

	storage := ctx.Value(StorageContextKey{}).(persistent.Storage)
	ads, err := storage.FindAdsWithTime(ctx, time.Now().UTC(), time.Now().UTC())
	require.NoError(t, err)
	require.Len(t, ads, len(testData))

	utils.SortAdsByID(ads)
	println(fmt.Sprint(ads))
	println(fmt.Sprint(testData))
	for _, ad := range ads {
		i := slices.IndexFunc(testData, func(ad2 models.Ad) bool {
			return ad.Title == ad2.Title
		})
		require.True(t, i >= 0, "ad not found", ad.Title)
		require.Equal(t, testData[i].StartAt, ad.StartAt)
		require.Equal(t, testData[i].EndAt, ad.EndAt)
	}
}
