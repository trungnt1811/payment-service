package set

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/genefriendway/onchain-handler/constants"
)

// TestItem represents a simple struct for testing Set functionality
type TestItem struct {
	ID   string
	Name string
}

// Key function for indexing
func testKeyFunc(item TestItem) string {
	return item.ID + "_" + item.Name
}

// Mock loader function
func mockLoader(ctx context.Context, limit, offset int) ([]TestItem, error) {
	items := []TestItem{}
	for i := offset; i < offset+limit; i++ {
		items = append(items, TestItem{ID: fmt.Sprintf("%d", i), Name: fmt.Sprintf("Item%d", i)})
	}
	return items, nil
}

// Test NewSet Initialization
func TestNewSet(t *testing.T) {
	t.Run("Create a new empty set", func(t *testing.T) {
		s, err := NewSet(context.Background(), testKeyFunc, mockLoader)
		require.NoError(t, err)
		require.NotNil(t, s)
		require.Equal(t, 0, len(s.GetAll()))
	})
}

// Test Adding Items
func TestAdd(t *testing.T) {
	t.Run("Add a unique item", func(t *testing.T) {
		s, _ := NewSet(context.Background(), testKeyFunc, mockLoader)
		item := TestItem{ID: "1", Name: "Alice"}

		require.NoError(t, s.Add(item))
		require.True(t, s.Contains(testKeyFunc(item)))
		require.Equal(t, 1, len(s.GetAll()))
	})

	t.Run("Fail to add duplicate item", func(t *testing.T) {
		s, _ := NewSet(context.Background(), testKeyFunc, mockLoader)
		item := TestItem{ID: "1", Name: "Alice"}

		require.NoError(t, s.Add(item))
		require.EqualError(t, s.Add(item), fmt.Sprintf("item %v already exists", item))
	})
}

// Test Removing Items
func TestRemove(t *testing.T) {
	t.Run("Remove an existing item", func(t *testing.T) {
		s, _ := NewSet(context.Background(), testKeyFunc, mockLoader)
		_ = s.Add(TestItem{ID: "1", Name: "Alice"})

		require.True(t, s.Remove(func(item TestItem) bool { return item.ID == "1" }))
		require.False(t, s.Contains("1_Alice"))
	})

	t.Run("Fail to remove non-existent item", func(t *testing.T) {
		s, _ := NewSet(context.Background(), testKeyFunc, mockLoader)

		require.False(t, s.Remove(func(item TestItem) bool { return item.ID == "99" }))
	})
}

// Test Retrieving Items
func TestGetItem(t *testing.T) {
	t.Run("Retrieve an existing item", func(t *testing.T) {
		s, _ := NewSet(context.Background(), testKeyFunc, mockLoader)
		_ = s.Add(TestItem{ID: "1", Name: "Alice"})

		item, exists := s.GetItem("1_Alice")
		require.True(t, exists)
		require.Equal(t, "Alice", item.Name)
	})

	t.Run("Fail to retrieve non-existent item", func(t *testing.T) {
		s, _ := NewSet(context.Background(), testKeyFunc, mockLoader)

		_, exists := s.GetItem("99_Unknown")
		require.False(t, exists)
	})
}

// Test Updating Items
func TestUpdateItem(t *testing.T) {
	t.Run("Update an existing item", func(t *testing.T) {
		s, _ := NewSet(context.Background(), testKeyFunc, mockLoader)
		_ = s.Add(TestItem{ID: "1", Name: "Alice"})

		newItem := TestItem{ID: "1", Name: "UpdatedAlice"}
		require.NoError(t, s.UpdateItem("1_Alice", newItem))

		updatedItem, exists := s.GetItem("1_Alice")
		require.True(t, exists)
		require.Equal(t, "UpdatedAlice", updatedItem.Name)
	})

	t.Run("Fail to update non-existent item", func(t *testing.T) {
		s, _ := NewSet(context.Background(), testKeyFunc, mockLoader)

		err := s.UpdateItem("99_Unknown", TestItem{ID: "99", Name: "DoesNotExist"})
		require.EqualError(t, err, "item with key 99_Unknown not found")
	})
}

// Test Filling Set
func TestFill(t *testing.T) {
	t.Run("Fill set with a batch of items", func(t *testing.T) {
		s, _ := NewSet(context.Background(), testKeyFunc, mockLoader)
		require.NoError(t, s.Fill(5))
		require.Equal(t, 5, len(s.GetAll()))
	})

	t.Run("Fill multiple times to accumulate items", func(t *testing.T) {
		s, _ := NewSet(context.Background(), testKeyFunc, mockLoader)

		require.NoError(t, s.Fill(3))
		require.NoError(t, s.Fill(2))
		require.Equal(t, 5, len(s.GetAll()))
	})

	t.Run("Fill should not exceed batch limit", func(t *testing.T) {
		s, _ := NewSet(context.Background(), testKeyFunc, mockLoader)

		require.NoError(t, s.Fill(5))
		require.Equal(t, 5, len(s.GetAll()))

		// Filling with limit=3 should add exactly 3 more items
		require.NoError(t, s.Fill(3))
		require.Equal(t, 8, len(s.GetAll()))
	})

	t.Run("Fail to fill when loader returns an error", func(t *testing.T) {
		faultyLoader := func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			return nil, fmt.Errorf("loader error")
		}

		s, _ := NewSet(context.Background(), testKeyFunc, faultyLoader)
		err := s.Fill(5)
		require.EqualError(t, err, "failed to load more items: loader error")
		require.Equal(t, 0, len(s.GetAll()))
	})
}

// Test Performance for Large Data
func TestSetPerformance(t *testing.T) {
	testPerformanceWithSize := func(size int) {
		s, _ := NewSet(context.Background(), testKeyFunc, mockLoader)

		for i := 0; i < size/constants.DefaultFillSetLimit; i++ { // Assume we fill in batches of 100
			err := s.Fill(constants.DefaultFillSetLimit)
			require.NoError(t, err)
		}

		require.Equal(t, size, len(s.GetAll()))
	}

	t.Run("Set Performance with 1M items", func(t *testing.T) {
		testPerformanceWithSize(1000000)
	})

	t.Run("Set Performance with 5M items", func(t *testing.T) {
		testPerformanceWithSize(5000000)
	})

	t.Run("Set Performance with 25M items", func(t *testing.T) {
		testPerformanceWithSize(25000000)
	})
}
