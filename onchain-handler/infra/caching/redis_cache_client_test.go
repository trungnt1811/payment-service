package caching

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/genefriendway/onchain-handler/conf"
)

func setupTestClient() *redisCacheClient {
	// Override the Redis configuration
	redisConfiguration := conf.GetRedisConfiguration()
	redisConfiguration.RedisAddress = "localhost:6379"
	redisConfiguration.RedisTtl = "5m"

	// Initialize the Redis cache client
	return NewRedisCacheClient().(*redisCacheClient)
}

func TestRedisCacheClient(t *testing.T) {
	client := setupTestClient()
	ctx := context.Background()

	// Table of test cases
	tests := []struct {
		name        string
		key         string
		value       interface{}
		dest        interface{}
		expiration  time.Duration
		expectError bool
	}{
		{"SetAndGet_String", "test_string_key", "test_value", new(string), 5 * time.Minute, false},
		{"SetAndGet_Int", "test_int_key", 42, new(int), 5 * time.Minute, false},
		{"SetAndGet_Bool", "test_bool_key", true, new(bool), 5 * time.Minute, false},
		{"SetAndGet_Struct", "test_struct_key", struct{ Name string }{"Alice"}, new(struct{ Name string }), 5 * time.Minute, false},
		{"Get_MissingKey", "non_existent_key", nil, new(string), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != nil {
				require.NoError(t, client.Set(ctx, tt.key, tt.value, tt.expiration))
			}

			err := client.Get(ctx, tt.key, tt.dest)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Compare values correctly
				switch v := tt.dest.(type) {
				case *int:
					require.Equal(t, tt.value.(int), *v)
				case *string:
					require.Equal(t, tt.value.(string), *v)
				case *bool:
					require.Equal(t, tt.value.(bool), *v)
				case *struct{ Name string }:
					require.Equal(t, tt.value.(struct{ Name string }).Name, v.Name)
				default:
					t.Fatalf("Unhandled type for test case: %s", tt.name)
				}
			}
		})
	}
}

func TestRedisCacheClient_Delete(t *testing.T) {
	client := setupTestClient()
	ctx := context.Background()

	t.Run("Delete_ExistingKey", func(t *testing.T) {
		key := "test_delete_key"
		value := "delete_me"
		expiration := 5 * time.Minute

		// Set a value
		require.NoError(t, client.Set(ctx, key, value, expiration))

		// Delete the value
		require.NoError(t, client.Del(ctx, key))

		// Ensure it's deleted
		var dest interface{}
		err := client.Get(ctx, key, &dest)
		require.Error(t, err)
		require.Equal(t, "cache miss for key: "+key, err.Error())
	})

	t.Run("Delete_NonExistingKey", func(t *testing.T) {
		key := "non_existent_key"
		require.NoError(t, client.Del(ctx, key))
	})
}
