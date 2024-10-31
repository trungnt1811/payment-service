package queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/genefriendway/onchain-handler/constants"
)

func TestNewQueue(t *testing.T) {
	loader := func(ctx context.Context, limit, offset int) ([]int, error) {
		return []int{1, 2, 3}, nil
	}

	q, err := NewQueue(context.Background(), 10, loader)
	require.NoError(t, err)
	require.Equal(t, 3, len(q.GetItems()))
}

func TestEnqueue(t *testing.T) {
	type NilableType struct {
		ID   int
		Name string
	}

	loader := func(ctx context.Context, limit, offset int) ([]*NilableType, error) {
		return []*NilableType{}, nil
	}

	q, err := NewQueue(context.Background(), 2, loader)
	require.NoError(t, err)

	// Add one item
	oneItem := &NilableType{ID: 1, Name: "Item1"}
	err = q.Enqueue(oneItem)
	require.NoError(t, err)

	// Check duplicate item
	err = q.Enqueue(oneItem)
	targetErrorMsg := fmt.Sprintf("item %v already exists in the queue", oneItem)
	require.EqualError(t, err, targetErrorMsg)

	// Check nil value
	err = q.Enqueue(nil)
	require.EqualError(t, err, "cannot enqueue a nil item")

	// Add another item
	twoItem := &NilableType{ID: 2, Name: "Item2"}
	err = q.Enqueue(twoItem)
	require.NoError(t, err)

	// Check queue is full
	threeItem := &NilableType{ID: 3, Name: "Item3"}
	err = q.Enqueue(threeItem)
	require.EqualError(t, err, "queue is full")
}

func TestContains(t *testing.T) {
	loader := func(ctx context.Context, limit, offset int) ([]int, error) {
		return []int{1, 2, 3}, nil
	}

	q, err := NewQueue(context.Background(), 10, loader)
	require.NoError(t, err)

	require.True(t, q.Contains(2))
	require.False(t, q.Contains(4))
}

func TestDequeue(t *testing.T) {
	loader := func(ctx context.Context, limit, offset int) ([]int, error) {
		return []int{1, 2, 3}, nil
	}

	q, err := NewQueue(context.Background(), 10, loader)
	require.NoError(t, err)

	err = q.Dequeue(func(item int) bool { return item == 2 })
	require.NoError(t, err)
	require.Equal(t, []int{1, 3}, q.GetItems())

	err = q.Dequeue(func(item int) bool { return item == 4 })
	require.Error(t, err)
}

func TestFillQueue(t *testing.T) {
	currentIndex := int64(0)
	loader := func(ctx context.Context, limit, offset int) ([]int64, error) {
		items := []int64{currentIndex, currentIndex + 1, currentIndex + 2}
		currentIndex += 3
		return items, nil
	}

	q, err := NewQueue(context.Background(), 6, loader)
	require.NoError(t, err)

	err = q.FillQueue()
	require.NoError(t, err)
	require.Equal(t, 6, len(q.GetItems()))
}

func TestGetSmallestValue(t *testing.T) {
	loader := func(ctx context.Context, limit, offset int) ([]int, error) {
		return []int{3, 1, 2}, nil
	}

	q, err := NewQueue(context.Background(), 10, loader)
	require.NoError(t, err)

	smallest, err := q.GetSmallestValue(func(a, b int) bool { return a < b })
	require.NoError(t, err)
	require.Equal(t, 1, smallest)
}

func TestGetItems(t *testing.T) {
	loader := func(ctx context.Context, limit, offset int) ([]int, error) {
		return []int{1, 2, 3, 4, 5}, nil
	}

	q, err := NewQueue(context.Background(), 10, loader)
	require.NoError(t, err)

	items := q.GetItems()
	require.Equal(t, 5, len(items))
	require.Equal(t, []int{1, 2, 3, 4, 5}, items)

	// Ensure that modifying the returned slice does not affect the queue's internal state
	items[0] = 100
	require.Equal(t, []int{1, 2, 3, 4, 5}, q.GetItems())
}

func TestGetLimit(t *testing.T) {
	loader := func(ctx context.Context, limit, offset int) ([]int, error) {
		return []int{1, 2, 3}, nil
	}

	q, err := NewQueue(context.Background(), 10, loader)
	require.NoError(t, err)

	require.Equal(t, 10, q.GetLimit())
}

func TestReplaceItemAtIndex(t *testing.T) {
	loader := func(ctx context.Context, limit, offset int) ([]int, error) {
		return []int{1, 2, 3, 4, 5}, nil
	}

	q, err := NewQueue(context.Background(), 10, loader)
	require.NoError(t, err)

	// Replace item at index 2
	err = q.ReplaceItemAtIndex(2, 10)
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 10, 4, 5}, q.GetItems())

	// Replace item at index 0
	err = q.ReplaceItemAtIndex(0, 20)
	require.NoError(t, err)
	require.Equal(t, []int{20, 2, 10, 4, 5}, q.GetItems())

	// Replace item at last index
	err = q.ReplaceItemAtIndex(4, 30)
	require.NoError(t, err)
	require.Equal(t, []int{20, 2, 10, 4, 30}, q.GetItems())

	// Check index out of bounds
	err = q.ReplaceItemAtIndex(5, 40)
	require.EqualError(t, err, "index out of bounds")
}

func TestQueuePerformance(t *testing.T) {
	testPerformanceWithSize := func(size int) {
		loader := func(ctx context.Context, limit, offset int) ([]int, error) {
			items := make([]int, limit)
			for i := 0; i < limit; i++ {
				items[i] = i + offset
			}
			return items, nil
		}

		q, err := NewQueue(context.Background(), 1, loader)
		require.NoError(t, err)

		start := time.Now()
		for i := 0; i < size; i++ {
			err := q.FillQueue()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}
		duration := time.Since(start)
		t.Logf("FillQueue %d items took %v", size, duration)

		start = time.Now()
		for i := 0; i < size; i++ {
			err := q.Dequeue(func(item int) bool { return item == i })
			if i > constants.MaxQueueLimit {
				require.Error(t, err)
				break
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}
		duration = time.Since(start)
		t.Logf("Dequeue %d items took %v", size, duration)
	}

	testPerformanceWithSize(500)    // 500
	testPerformanceWithSize(1000)   // 1K
	testPerformanceWithSize(5000)   // 5K
	testPerformanceWithSize(10000)  // 10K
	testPerformanceWithSize(100000) // 100K
	// testPerformanceWithSize(1000000) // 1M
}
