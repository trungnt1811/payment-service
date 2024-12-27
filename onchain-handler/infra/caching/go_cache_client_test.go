package caching

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewGoCacheClient(t *testing.T) {
	t.Run("NewGoCacheClient", func(t *testing.T) {
		client := NewGoCacheClient()
		require.NotNil(t, client)
	})
}

func TestGoCacheClient_Set(t *testing.T) {
	t.Run("GoCacheClient_Set", func(t *testing.T) {
		client := NewGoCacheClient()
		ctx := context.Background()
		key := "GoCacheClient_Set_Key"
		value := "GoCacheClient_Set_Value"
		expiration := 5 * time.Minute

		setErr := client.Set(ctx, key, value, expiration)
		require.NoError(t, setErr)

		dest := ""
		getErr := client.Get(ctx, key, &dest)
		require.NoError(t, getErr)
		require.Equal(t, value, dest)
	})
}

func TestGoCacheClient_Get(t *testing.T) {
	t.Run("GoCacheClient_Get", func(t *testing.T) {
		client := NewGoCacheClient()
		ctx := context.Background()
		key := "GoCacheClient_Get_Key"
		value := "GoCacheClient_Get_Value"
		expiration := 5 * time.Minute

		setErr := client.Set(ctx, key, value, expiration)
		require.NoError(t, setErr)

		dest := ""
		getErr := client.Get(ctx, key, &dest)
		require.NoError(t, getErr)
		require.Equal(t, value, dest)
	})

	t.Run("GoCacheClient_Get_ItemNotFound", func(t *testing.T) {
		client := NewGoCacheClient()
		ctx := context.Background()
		key := "nonExistentKey"

		dest := ""
		getErr := client.Get(ctx, key, &dest)
		require.Error(t, getErr)
		require.Equal(t, "item not found in cache", getErr.Error())
	})

	t.Run("GoCacheClient_Get_InvalidDestination", func(t *testing.T) {
		client := NewGoCacheClient()
		ctx := context.Background()
		key := "GoCacheClient_Get_InvalidDestination_Key"
		value := "GoCacheClient_Get_InvalidDestination_Value"
		expiration := 5 * time.Minute

		setErr := client.Set(ctx, key, value, expiration)
		require.NoError(t, setErr)

		dest := 0
		getErr := client.Get(ctx, key, &dest)
		require.Error(t, getErr)
		require.Equal(t, "cached value type (string) does not match destination type (int)", getErr.Error())
	})

	t.Run("GoCacheClient_Get_NilDestination", func(t *testing.T) {
		client := NewGoCacheClient()
		ctx := context.Background()
		key := "GoCacheClient_Get_NilDestination_Key"
		value := "GoCacheClient_Get_NilDestination_Value"
		expiration := 5 * time.Minute

		setErr := client.Set(ctx, key, value, expiration)
		require.NoError(t, setErr)

		var dest *string
		getErr := client.Get(ctx, key, dest)
		require.Error(t, getErr)
		require.Equal(t, "destination must be a non-nil pointer", getErr.Error())
	})
}

func TestGoCacheClient_Del(t *testing.T) {
	t.Run("GoCacheClient_Del", func(t *testing.T) {
		client := NewGoCacheClient()
		ctx := context.Background()
		key := "GoCacheClient_Del_Key"
		value := "GoCacheClient_Del_Value"
		expiration := 5 * time.Minute

		setErr := client.Set(ctx, key, value, expiration)
		require.NoError(t, setErr)

		delErr := client.Del(ctx, key)
		require.NoError(t, delErr)

		dest := ""
		getErr := client.Get(ctx, key, &dest)
		require.Equal(t, "item not found in cache", getErr.Error())
		require.Equal(t, "", dest)
	})
}
