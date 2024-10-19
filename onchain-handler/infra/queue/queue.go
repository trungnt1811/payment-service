package queue

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/genefriendway/onchain-handler/constants"
)

type Queue[T comparable] struct {
	ctx     context.Context
	mu      sync.Mutex
	items   []T
	itemSet map[T]struct{}
	limit   int
	loader  func(ctx context.Context, limit, offset int) ([]T, error)
}

// NewQueue creates a new queue with an initial load of items.
func NewQueue[T comparable](ctx context.Context, limit int, loader func(ctx context.Context, limit, offset int) ([]T, error)) (*Queue[T], error) {
	q := &Queue[T]{
		ctx:     ctx,
		items:   make([]T, 0, limit),
		itemSet: make(map[T]struct{}),
		limit:   limit,
		loader:  loader,
	}

	if err := q.loadInitialItems(); err != nil {
		return nil, err
	}
	return q, nil
}

// loadInitialItems populates the queue with the first set of items.
func (q *Queue[T]) loadInitialItems() error {
	items, err := q.loader(q.ctx, q.limit, 0)
	if err != nil {
		return fmt.Errorf("failed to load initial items: %w", err)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range items {
		q.items = append(q.items, item)
		q.itemSet[item] = struct{}{}
	}
	return nil
}

// Enqueue adds a new item, ensuring no duplicates and enforcing the limit.
func (q *Queue[T]) Enqueue(item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.itemSet[item]; exists {
		return fmt.Errorf("item %v already exists in the queue", item)
	}

	if len(q.items) >= q.limit {
		return fmt.Errorf("queue is full")
	}

	if reflect.ValueOf(item).IsNil() {
		return fmt.Errorf("cannot enqueue a nil item")
	}

	q.items = append(q.items, item)
	q.itemSet[item] = struct{}{}
	return nil
}

// Dequeue removes the first item that matches the condition.
func (q *Queue[T]) Dequeue(condition func(T) bool) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, item := range q.items {
		if condition(item) {
			log.Printf("Dequeueing item: %v", item)
			// Fast removal, maintaining order
			copy(q.items[i:], q.items[i+1:])
			q.items = q.items[:len(q.items)-1]
			delete(q.itemSet, item)

			// Check if shrinking is necessary
			if q.limit > constants.MinQueueLimit {
				q.maybeShrink()
			}
			return nil
		}
	}
	return fmt.Errorf("item not found")
}

// FillQueue loads more items to fill up to the limit.
func (q *Queue[T]) FillQueue() error {
	q.mu.Lock()
	offset := len(q.items)
	remainingCapacity := q.limit - offset
	q.mu.Unlock()

	// Load more items
	newItems, err := q.loader(q.ctx, remainingCapacity+1, offset)
	if err != nil {
		log.Printf("Error loading items: %v", err)
		return fmt.Errorf("failed to load more items: %w", err)
	}
	log.Printf("Loaded %d new items", len(newItems))

	q.mu.Lock()
	defer q.mu.Unlock()

	// Scale if necessary
	if len(newItems) > remainingCapacity {
		q.scaleQueue()
	}

	// Add unique new items
	for _, item := range newItems {
		if _, exists := q.itemSet[item]; !exists {
			q.items = append(q.items, item)
			q.itemSet[item] = struct{}{}
		}
	}
	return nil
}

// scaleQueue scales the queue when needed.
func (q *Queue[T]) scaleQueue() {
	newLimit := int(float64(q.limit) * constants.ScaleFactor)
	if newLimit > constants.MaxQueueLimit {
		q.limit = constants.MaxQueueLimit
	} else {
		q.limit = newLimit
	}
	log.Printf("Scaling queue to new limit: %d", q.limit)
}

// maybeShrink reduces the queue's size if it's consistently underutilized.
func (q *Queue[T]) maybeShrink() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= int(float64(q.limit)*constants.ShrinkThreshold) {
		return
	}

	if q.limit > constants.MinQueueLimit {
		newLimit := int(float64(q.limit) * constants.ShrinkFactor)
		if newLimit < constants.MinQueueLimit {
			q.limit = constants.MinQueueLimit
		} else {
			q.limit = newLimit
		}
	}
	log.Printf("Shrinking queue to new limit: %d", q.limit)
}

// GetItems returns a copy of all current items.
func (q *Queue[T]) GetItems() []T {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.items
}

// GetSmallestValue finds the smallest value according to a comparator function.
func (q *Queue[T]) GetSmallestValue(compare func(T, T) bool) (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		var zero T
		return zero, fmt.Errorf("queue is empty")
	}

	smallest := q.items[0]
	for _, item := range q.items[1:] {
		if compare(item, smallest) {
			smallest = item
		}
	}

	return smallest, nil
}

func (q *Queue[T]) GetLimit() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.limit
}
