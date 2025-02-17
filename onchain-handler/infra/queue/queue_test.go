package queue

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/genefriendway/onchain-handler/pkg/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

// Initialize the logger once for all tests
func TestMain(m *testing.M) {
	logger.SetLogLevel(interfaces.DebugLevel)
	logger.GetLogger().Info("Starting test suite for queue package")
	os.Exit(m.Run())
}

// TestItem represents a simple struct for testing queue functionality
type TestItem struct {
	ID   string
	Name string
}

// Key function for indexing
func testKeyFunc(item TestItem) string {
	return item.ID + "_" + item.Name
}

// Test Queue Initialization
func TestNewQueue(t *testing.T) {
	t.Run("Initialize Queue with Loader", func(t *testing.T) {
		loader := func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			return []TestItem{{ID: "1", Name: "Alice"}, {ID: "2", Name: "Bob"}}, nil
		}

		q, err := NewQueue(context.Background(), 10, testKeyFunc, loader)
		require.NoError(t, err)
		require.Equal(t, 2, len(q.GetItems()))
	})
}

// Test Enqueueing Items
func TestEnqueue(t *testing.T) {
	t.Run("Enqueue Unique Item", func(t *testing.T) {
		q, _ := NewQueue(context.Background(), 3, testKeyFunc, func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			return []TestItem{}, nil
		})

		item := TestItem{ID: "1", Name: "Item1"}
		require.NoError(t, q.Enqueue(item))
		require.Equal(t, 1, len(q.GetItems()))
	})

	t.Run("Check Duplicate Item", func(t *testing.T) {
		q, _ := NewQueue(context.Background(), 2, testKeyFunc, func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			return []TestItem{}, nil
		})

		item := TestItem{ID: "1", Name: "Item1"}
		require.NoError(t, q.Enqueue(item))

		err := q.Enqueue(item)
		require.EqualError(t, err, fmt.Sprintf("item %v already exists in the queue", item))
	})

	t.Run("Queue Full Error", func(t *testing.T) {
		q, _ := NewQueue(context.Background(), 2, testKeyFunc, func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			return []TestItem{}, nil
		})

		require.NoError(t, q.Enqueue(TestItem{ID: "1", Name: "Item1"}))
		require.NoError(t, q.Enqueue(TestItem{ID: "2", Name: "Item2"}))

		err := q.Enqueue(TestItem{ID: "3", Name: "Item3"})
		require.EqualError(t, err, "queue is full")
	})
}

// Test Dequeueing Items
func TestDequeue(t *testing.T) {
	t.Run("Dequeue Item", func(t *testing.T) {
		loader := func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			return []TestItem{{ID: "1", Name: "Alice"}, {ID: "2", Name: "Bob"}, {ID: "3", Name: "Charlie"}}, nil
		}

		q, _ := NewQueue(context.Background(), 10, testKeyFunc, loader)

		err := q.Dequeue(func(item TestItem) bool { return item.ID == "2" })
		require.NoError(t, err)
		require.Equal(t, 2, len(q.GetItems()))
	})

	t.Run("Dequeue Non-Existent Item", func(t *testing.T) {
		q, _ := NewQueue(context.Background(), 10, testKeyFunc, func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			return []TestItem{{ID: "1", Name: "Alice"}}, nil
		})

		err := q.Dequeue(func(item TestItem) bool { return item.ID == "99" })
		require.Error(t, err)
	})
}

// Test Filling Queue
func TestFillQueue(t *testing.T) {
	t.Run("Fill Queue to Limit", func(t *testing.T) {
		currentIndex := int64(0)
		loader := func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			items := []TestItem{
				{ID: fmt.Sprintf("%d", currentIndex), Name: "Alice"},
				{ID: fmt.Sprintf("%d", currentIndex+1), Name: "Bob"},
				{ID: fmt.Sprintf("%d", currentIndex+2), Name: "Timmy"},
			}
			currentIndex += 3
			return items, nil
		}

		q, err := NewQueue(context.Background(), 6, testKeyFunc, loader)
		require.NoError(t, err)

		err = q.FillQueue()
		require.NoError(t, err)
		require.Equal(t, 6, len(q.GetItems()))
	})
}

// Test Index Retrieval
func TestGetIndex(t *testing.T) {
	t.Run("Retrieve Index of an Item", func(t *testing.T) {
		q, _ := NewQueue(context.Background(), 10, testKeyFunc, func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			return []TestItem{{ID: "1", Name: "Alice"}, {ID: "2", Name: "Bob"}}, nil
		})

		index, exists := q.GetIndex("2_Bob")
		require.True(t, exists)
		require.Equal(t, 1, index)

		_, exists = q.GetIndex("99_Unknown")
		require.False(t, exists)
	})
}

// Test Replacing Items
func TestReplaceItemAtIndex(t *testing.T) {
	t.Run("Replace Items at Various Indexes", func(t *testing.T) {
		q, _ := NewQueue(context.Background(), 10, testKeyFunc, func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			return []TestItem{{ID: "1", Name: "Alice"}, {ID: "2", Name: "Bob"}, {ID: "3", Name: "Charlie"}}, nil
		})

		require.NoError(t, q.ReplaceItemAtIndex(1, TestItem{ID: "4", Name: "David"}))
		require.Equal(t, "David", q.GetItems()[1].Name)

		err := q.ReplaceItemAtIndex(5, TestItem{ID: "5", Name: "Out of Bounds"})
		require.EqualError(t, err, "index out of bounds")
	})
}

func TestGetItemAtIndex(t *testing.T) {
	t.Run("Retrieve Item at Valid Index", func(t *testing.T) {
		q, _ := NewQueue(context.Background(), 5, testKeyFunc, func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			return []TestItem{
				{ID: "1", Name: "Alice"},
				{ID: "2", Name: "Bob"},
				{ID: "3", Name: "Charlie"},
			}, nil
		})

		// Get item at index 1 (Bob)
		item, err := q.GetItemAtIndex(1)
		require.NoError(t, err)
		require.Equal(t, "2", item.ID)
		require.Equal(t, "Bob", item.Name)
	})

	t.Run("Retrieve Item at Out of Bounds Index", func(t *testing.T) {
		q, _ := NewQueue(context.Background(), 3, testKeyFunc, func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			return []TestItem{
				{ID: "1", Name: "Alice"},
				{ID: "2", Name: "Bob"},
			}, nil
		})

		_, err := q.GetItemAtIndex(5)
		require.EqualError(t, err, "index out of bounds")
	})

	t.Run("Retrieve Copy of Item (Modify Without Affecting Queue)", func(t *testing.T) {
		q, _ := NewQueue(context.Background(), 3, testKeyFunc, func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			return []TestItem{
				{ID: "1", Name: "Alice"},
				{ID: "2", Name: "Bob"},
			}, nil
		})

		// Get item at index 0
		item, err := q.GetItemAtIndex(0)
		require.NoError(t, err)

		// Modify the retrieved item
		item.Name = "ModifiedAlice"
		require.Equal(t, "ModifiedAlice", item.Name, "Queue item should be modified outside the queue")

		// Get item again from the queue to ensure it's unchanged
		originalItem, _ := q.GetItemAtIndex(0)
		require.Equal(t, "Alice", originalItem.Name, "Queue item should not be modified externally")
	})
}

// Test Performance
func TestQueuePerformance(t *testing.T) {
	testPerformanceWithSize := func(size int) {
		loader := func(ctx context.Context, limit, offset int) ([]TestItem, error) {
			items := make([]TestItem, limit)
			for i := 0; i < limit; i++ {
				items[i] = TestItem{ID: fmt.Sprintf("%d", i+offset), Name: fmt.Sprintf("Item%d", i)}
			}
			return items, nil
		}

		q, _ := NewQueue(context.Background(), 100, testKeyFunc, loader)

		start := time.Now()
		for i := 0; i < size; i++ {
			err := q.FillQueue()
			require.NoError(t, err, "FillQueue failed at iteration %d", i)
		}
		t.Logf("FillQueue %d items took %v", size, time.Since(start))
	}

	t.Run("Queue Performance with 10K items", func(t *testing.T) {
		testPerformanceWithSize(10000)
	})
	t.Run("Queue Performance with 50K items", func(t *testing.T) {
		testPerformanceWithSize(50000)
	})
	t.Run("Queue Performance with 250K items", func(t *testing.T) {
		testPerformanceWithSize(250000)
	})
}
