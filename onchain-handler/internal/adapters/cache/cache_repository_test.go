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

		prefixedKey := config.AppName + "_testKey"
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

		prefixedKey := config.AppName + "_testKey"
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

		prefixedKey := config.AppName + "_testKey"

		mockClient.EXPECT().Del(ctx, prefixedKey).Return(nil)

		err := repo.RemoveItem(mockStringer)
		require.NoError(t, err)
	})
}
