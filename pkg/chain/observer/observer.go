package observer

import (
	"chainnet/pkg/kernel"
	"sync"
)

// Observer interface that defines the methods that an observer should implement
type Observer interface {
	ID() string
	OnBlockAddition(block *kernel.Block)
}

// Subject controller that manages the observers
type Subject interface {
	Register(observer Observer)
	Unregister(observer Observer)
	NotifyBlockAdded(block *kernel.Block)
}

type SubjectObserver struct {
	observers map[string]Observer
	mu        sync.Mutex
}

func NewSubjectObserver() *SubjectObserver {
	return &SubjectObserver{
		observers: make(map[string]Observer),
	}
}

// Register adds an observer to the list of observers
func (so *SubjectObserver) Register(observer Observer) {
	so.mu.Lock()
	defer so.mu.Unlock()
	so.observers[observer.ID()] = observer
}

// Unregister removes an observer from the list of observers
func (so *SubjectObserver) Unregister(observer Observer) {
	so.mu.Lock()
	defer so.mu.Unlock()
	delete(so.observers, observer.ID())
}

// NotifyBlockAdded notifies all observers that a new block has been added
func (so *SubjectObserver) NotifyBlockAdded(block *kernel.Block) {
	so.mu.Lock()
	defer so.mu.Unlock()
	for _, observer := range so.observers {
		go observer.OnBlockAddition(block)
	}
}
