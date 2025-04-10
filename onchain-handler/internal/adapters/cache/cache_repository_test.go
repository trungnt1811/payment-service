package cache

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/require"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/adapters/cache/mocks"
)

func buildPrefixedKey(appName, key string) string {
	return appName + "_" + key
}

func TestSaveItem(t *testing.T) {
	t.Run("TestCachingRepository_SaveItem", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCacheClient := mocks.NewMockCacheClient(ctrl)
		mockStringer := mocks.NewMockStringer(ctrl)

		ctx := context.Background()
		config := &conf.Configuration{AppName: conf.GetAppName()} // Ensure config consistency
		repo := NewCachingRepository(ctx, mockCacheClient)

		// Define mock key expectations properly
		mockStringer.EXPECT().String().Return("testKey").AnyTimes()

		prefixedKey := buildPrefixedKey(config.AppName, "testKey")
		value := "testValue"
		expire := 5 * time.Minute

		// Expectation on mocked cache client
		mockCacheClient.EXPECT().Set(ctx, prefixedKey, value, expire).Return(nil)

		err := repo.SaveItem(mockStringer, value, expire)
		require.NoError(t, err)
	})
}

func TestRetrieveItem(t *testing.T) {
	t.Run("TestCachingRepository_RetrieveItem", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockCacheClient(ctrl)
		mockStringer := mocks.NewMockStringer(ctrl)

		ctx := context.Background()
		config := &conf.Configuration{AppName: conf.GetAppName()} // Ensure consistent config usage
		repo := NewCachingRepository(ctx, mockClient)

		mockStringer.EXPECT().String().Return("testKey").AnyTimes()

		prefixedKey := buildPrefixedKey(config.AppName, "testKey")
		var value string

		mockClient.EXPECT().Get(ctx, prefixedKey, &value).Return(nil)

		err := repo.RetrieveItem(mockStringer, &value)
		require.NoError(t, err)
	})
}

func TestRemoveItem(t *testing.T) {
	t.Run("TestCachingRepository_RemoveItem", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockCacheClient(ctrl)
		mockStringer := mocks.NewMockStringer(ctrl)

		ctx := context.Background()
		config := &conf.Configuration{AppName: conf.GetAppName()} // Ensure consistent config usage
		repo := NewCachingRepository(ctx, mockClient)

		mockStringer.EXPECT().String().Return("testKey").AnyTimes()

		prefixedKey := buildPrefixedKey(config.AppName, "testKey")

		mockClient.EXPECT().Del(ctx, prefixedKey).Return(nil)

		err := repo.RemoveItem(mockStringer)
		require.NoError(t, err)
	})
}

func TestGetAllMatching(t *testing.T) {
	t.Run("TestCachingRepository_GetAllMatching", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockCacheClient(ctrl)
		mockPrefix := mocks.NewMockStringer(ctrl)

		ctx := context.Background()
		config := &conf.Configuration{AppName: conf.GetAppName()}
		repo := NewCachingRepository(ctx, mockClient)

		mockPrefix.EXPECT().String().Return("testPrefix").AnyTimes()

		fullPrefix := buildPrefixedKey(config.AppName, "testPrefix")

		type Dummy struct {
			Value string
		}

		expectedResults := []any{
			&Dummy{Value: "One"},
			&Dummy{Value: "Two"},
		}

		// Expect the client to return those dummy values
		mockClient.EXPECT().
			GetAllMatching(ctx, fullPrefix, gomock.Any()).
			Return(expectedResults, nil)

		results, err := repo.GetAllMatching(mockPrefix, func() any {
			return new(Dummy)
		})

		require.NoError(t, err)
		require.Len(t, results, 2)

		v1 := results[0].(*Dummy)
		v2 := results[1].(*Dummy)

		require.Equal(t, "One", v1.Value)
		require.Equal(t, "Two", v2.Value)
	})
}
