package types

import "context"

type Set[T comparable] interface {
	GetAll() []T                                              // Returns all items in the set
	Contains(key string) bool                                 // Checks if an item exists in the set
	GetItem(key string) (T, bool)                             // Retrieves an item by key if it exists
	Add(item T) error                                         // Adds a new item to the set
	Remove(condition func(T) bool) bool                       // Removes an item from the set based on a condition
	UpdateItem(key string, newItem T) error                   // Updates an item by key
	Fill(loader func(ctx context.Context) ([]T, error)) error // Ensures the set is populated
}
