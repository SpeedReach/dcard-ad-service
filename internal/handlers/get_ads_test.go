package handlers

import (
	"advertise_service/internal/infra/persistent"
	"advertise_service/internal/mock"
	"advertise_service/internal/models"
	"advertise_service/internal/utils"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestParseRequest(t *testing.T) {
	request, err := http.NewRequest("GET", "/ad?limit=3&offset=5&age=24&gender=F&country=TW&platform=ios", nil)
	assert.NoError(t, err)
	req, err := ParseGetAdsRequest(request)
	require.NoError(t, err)
	assert.Equal(t, 3, req.Limit)
	assert.Equal(t, 5, req.Offset)
	assert.Equal(t, 24, req.Age)
	assert.Equal(t, models.Female, req.Gender)
	assert.Equal(t, models.Taiwan, req.Country)
	assert.Equal(t, models.Ios, req.Platform)
}

func TestGetAd(t *testing.T) {
	testData := mock.GenerateMockAds()
	ctx := InjectMockedResources(context.Background())
	storage := ctx.Value(StorageContextKey{}).(persistent.Storage)
	for _, ad := range testData {
		err := storage.InsertAd(ctx, ad)
		require.NoError(t, err)
	}

	getAd := func(t *testing.T) {
		request := GetAdsRequest{
			Offset:   0,
			Limit:    1000,
			Age:      24,
			Gender:   models.Male,
			Country:  models.Japan,
			Platform: models.Web,
		}

		response, err := fetchMatched(ctx, request)
		assert.NoError(t, err)
		utils.SortAdsByEndTimeAsc(testData)

		condParam := ExtractConditionParams(request)
		i := 0
		for _, testAd := range testData {
			if testAd.ShouldShow(condParam) {
				require.True(t, i < len(response.Items), "out of range", i, len(response.Items))
				assert.Equal(t, testAd.Title, response.Items[i].Title)
				assert.True(t, mock.MockedAdShouldShow(testAd, condParam))
				i++
			}
		}
	}

	t.Run("GetAdFromStorage", func(t *testing.T) {
		getAd(t)
	})

	t.Run("GetAdFromCache", func(t *testing.T) {
		getAd(t)
	})

}
