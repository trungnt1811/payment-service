package queue

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/pkg/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

// Initialize the logger once for all tests
func TestMain(m *testing.M) {
	// Initialize the logger with a debug log level for testing
	logger.SetLogLevel(interfaces.DebugLevel)
	appLogger := logger.GetLogger()

	// Set service details for consistent context in logs
	appLogger.Info("Starting test suite for my-service in development mode")

	// Run all tests
	code := m.Run()

	// Exit with the code from test execution
	os.Exit(code)
}

func TestNewQueue(t *testing.T) {
	t.Run("Initialize Queue with Loader", func(t *testing.T) {
		loader := func(ctx context.Context, limit, offset int) ([]int, error) {
			return []int{1, 2, 3}, nil
		}

		q, err := NewQueue(context.Background(), 10, loader)
		require.NoError(t, err)
		require.Equal(t, 3, len(q.GetItems()))
	})
}

func TestEnqueue(t *testing.T) {
	type NilableType struct {
		ID   int
		Name string
	}

	t.Run("Enqueue Item", func(t *testing.T) {
		loader := func(ctx context.Context, limit, offset int) ([]*NilableType, error) {
			return []*NilableType{}, nil
		}

		q, err := NewQueue(context.Background(), 2, loader)
		require.NoError(t, err)

		oneItem := &NilableType{ID: 1, Name: "Item1"}
		err = q.Enqueue(oneItem)
		require.NoError(t, err)
	})

	t.Run("Check Duplicate Item", func(t *testing.T) {
		q, err := NewQueue(context.Background(), 2, func(ctx context.Context, limit, offset int) ([]*NilableType, error) {
			return []*NilableType{}, nil
		})
		require.NoError(t, err)

		oneItem := &NilableType{ID: 1, Name: "Item1"}
		err = q.Enqueue(oneItem)
		require.NoError(t, err)

		err = q.Enqueue(oneItem)
		targetErrorMsg := fmt.Sprintf("item %v already exists in the queue", oneItem)
		require.EqualError(t, err, targetErrorMsg)
	})

	t.Run("Enqueue Nil Value", func(t *testing.T) {
		q, _ := NewQueue(context.Background(), 2, func(ctx context.Context, limit, offset int) ([]*NilableType, error) {
			return []*NilableType{}, nil
		})
		err := q.Enqueue(nil)
		require.EqualError(t, err, "cannot enqueue a nil item")
	})

	t.Run("Queue Full Error", func(t *testing.T) {
		q, _ := NewQueue(context.Background(), 2, func(ctx context.Context, limit, offset int) ([]*NilableType, error) {
			return []*NilableType{}, nil
		})

		var err error
		err = q.Enqueue(&NilableType{ID: 1, Name: "Item1"})
		require.NoError(t, err)
		err = q.Enqueue(&NilableType{ID: 2, Name: "Item2"})
		require.NoError(t, err)

		threeItem := &NilableType{ID: 3, Name: "Item3"}
		err = q.Enqueue(threeItem)
		require.EqualError(t, err, "queue is full")
	})
}

func TestDequeue(t *testing.T) {
	t.Run("Dequeue Item", func(t *testing.T) {
		loader := func(ctx context.Context, limit, offset int) ([]int, error) {
			return []int{1, 2, 3}, nil
		}

		q, err := NewQueue(context.Background(), 10, loader)
		require.NoError(t, err)

		err = q.Dequeue(func(item int) bool { return item == 2 })
		require.NoError(t, err)
		require.Equal(t, []int{1, 3}, q.GetItems())
	})

	t.Run("Dequeue Non-Existent Item", func(t *testing.T) {
		loader := func(ctx context.Context, limit, offset int) ([]int, error) {
			return []int{1, 2, 3}, nil
		}

		q, err := NewQueue(context.Background(), 10, loader)
		require.NoError(t, err)

		err = q.Dequeue(func(item int) bool { return item == 4 })
		require.Error(t, err)
	})
}

func TestFillQueue(t *testing.T) {
	t.Run("Fill Queue to Limit", func(t *testing.T) {
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
	})
}

func TestGetItems(t *testing.T) {
	t.Run("Retrieve All Items", func(t *testing.T) {
		loader := func(ctx context.Context, limit, offset int) ([]int, error) {
			return []int{1, 2, 3, 4, 5}, nil
		}

		q, err := NewQueue(context.Background(), 10, loader)
		require.NoError(t, err)

		items := q.GetItems()
		require.Equal(t, 5, len(items))
		require.Equal(t, []int{1, 2, 3, 4, 5}, items)

		items[0] = 100
		require.Equal(t, []int{1, 2, 3, 4, 5}, q.GetItems())
	})
}

func TestReplaceItemAtIndex(t *testing.T) {
	t.Run("Replace Items at Various Indexes", func(t *testing.T) {
		loader := func(ctx context.Context, limit, offset int) ([]int, error) {
			return []int{1, 2, 3, 4, 5}, nil
		}

		q, err := NewQueue(context.Background(), 10, loader)
		require.NoError(t, err)

		err = q.ReplaceItemAtIndex(2, 10)
		require.NoError(t, err)
		require.Equal(t, []int{1, 2, 10, 4, 5}, q.GetItems())

		err = q.ReplaceItemAtIndex(0, 20)
		require.NoError(t, err)
		require.Equal(t, []int{20, 2, 10, 4, 5}, q.GetItems())

		err = q.ReplaceItemAtIndex(4, 30)
		require.NoError(t, err)
		require.Equal(t, []int{20, 2, 10, 4, 30}, q.GetItems())

		err = q.ReplaceItemAtIndex(5, 40)
		require.EqualError(t, err, "index out of bounds")
	})
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
			require.NoError(t, err)
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

	t.Run("Queue Performance with 1K items", func(t *testing.T) {
		testPerformanceWithSize(1000)
	})
	t.Run("Queue Performance with 10K items", func(t *testing.T) {
		testPerformanceWithSize(10000)
	})
	t.Run("Queue Performance with 1M items", func(t *testing.T) {
		testPerformanceWithSize(1000000)
	})
}
