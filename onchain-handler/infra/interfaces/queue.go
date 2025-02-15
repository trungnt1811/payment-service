package interfaces

type Queue[T comparable] interface {
	GetItems() []T
	FillQueue() error
	Enqueue(item T) error
	Dequeue(condition func(T) bool) error
	ReplaceItemAtIndex(index int, newItem T) error
}
