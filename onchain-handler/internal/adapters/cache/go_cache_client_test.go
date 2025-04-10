package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
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
		require.True(t, errors.Is(getErr, types.ErrNotFound))
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
		require.True(t, errors.Is(getErr, types.ErrNotFound))
		require.Equal(t, "", dest)
	})
}

func TestGoCacheClient_GetAllMatching(t *testing.T) {
	t.Run("GoCacheClient_GetAllMatching", func(t *testing.T) {
		client := NewGoCacheClient()
		ctx := context.Background()

		// Setup
		type TestStruct struct {
			Name string
		}

		// Save multiple items with a common prefix
		prefix := "GoCacheClient_GetAllMatching_"
		items := map[string]TestStruct{
			prefix + "A":   {Name: "Alice"},
			prefix + "B":   {Name: "Bob"},
			"UnrelatedKey": {Name: "Eve"},
		}

		for key, val := range items {
			err := client.Set(ctx, key, val, 5*time.Minute)
			require.NoError(t, err)
		}

		// Fetch only those with matching prefix
		results, err := client.GetAllMatching(ctx, prefix, func() any {
			return new(TestStruct)
		})
		require.NoError(t, err)
		require.Len(t, results, 2)

		names := make(map[string]bool)
		for _, raw := range results {
			ts, ok := raw.(*TestStruct)
			require.True(t, ok)
			names[ts.Name] = true
		}

		require.True(t, names["Alice"])
		require.True(t, names["Bob"])
		require.False(t, names["Eve"])
	})
}
