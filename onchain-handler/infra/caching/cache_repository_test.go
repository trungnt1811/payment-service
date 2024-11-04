package caching

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/require"

	"github.com/genefriendway/onchain-handler/infra/interfaces"
)

func TestSaveItem(t *testing.T) {
	t.Run("TestCachingRepository_SaveItem", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCacheClient := interfaces.NewMockCacheClient(ctrl)
		ctx := context.Background()
		repo := NewCachingRepository(ctx, mockCacheClient)

		key := interfaces.NewMockStringer(ctrl).EXPECT().String().Return("testKey").AnyTimes()
		value := "testValue"
		expire := 5 * time.Minute

		mockCacheClient.EXPECT().Set(ctx, key.String(), value, expire).Return(nil)

		err := repo.SaveItem(key, value, expire)
		require.NoError(t, err)
	})
}

func TestRetrieveItem(t *testing.T) {
	t.Run("TestCachingRepository_RetrieveItem", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := interfaces.NewMockCacheClient(ctrl)
		ctx := context.Background()
		repo := NewCachingRepository(ctx, mockClient)

		key := interfaces.NewMockStringer(ctrl).EXPECT().String().Return("testKey").AnyTimes()
		var value string

		mockClient.EXPECT().Get(ctx, key.String(), &value).Return(nil)

		err := repo.RetrieveItem(key, &value)
		require.NoError(t, err)
	})
}

func TestRemoveItem(t *testing.T) {
	t.Run("TestCachingRepository_RemoveItem", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := interfaces.NewMockCacheClient(ctrl)
		ctx := context.Background()
		repo := NewCachingRepository(ctx, mockClient)

		key := interfaces.NewMockStringer(ctrl).EXPECT().String().Return("testKey").AnyTimes()

		mockClient.EXPECT().Del(ctx, key.String()).Return(nil)

		err := repo.RemoveItem(key)
		require.NoError(t, err)
	})
}
