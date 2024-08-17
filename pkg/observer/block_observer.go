package observer

import (
	"chainnet/pkg/kernel"
	"sync"
)

// BlockObserver interface that defines the methods that a block observer should implement
type BlockObserver interface {
	BlockObserverID() string
	OnBlockAddition(block *kernel.Block)
}

// BlockSubject controller that manages the block observers
type BlockSubject interface {
	Register(observer BlockObserver)
	Unregister(observer BlockObserver)
	NotifyBlockAdded(block *kernel.Block)
}

type BlockSubjectController struct {
	observers map[string]BlockObserver
	mu        sync.Mutex
}

func NewBlockSubject() *BlockSubjectController {
	return &BlockSubjectController{
		observers: make(map[string]BlockObserver),
	}
}

// Register adds an observer to the list of observers
func (so *BlockSubjectController) Register(observer BlockObserver) {
	so.mu.Lock()
	defer so.mu.Unlock()
	so.observers[observer.BlockObserverID()] = observer
}

// Unregister removes an observer from the list of observers
func (so *BlockSubjectController) Unregister(observer BlockObserver) {
	so.mu.Lock()
	defer so.mu.Unlock()
	delete(so.observers, observer.BlockObserverID())
}

// NotifyBlockAdded notifies all observers that a new block has been added
func (so *BlockSubjectController) NotifyBlockAdded(block *kernel.Block) {
	so.mu.Lock()
	defer so.mu.Unlock()
	for _, observer := range so.observers {
		observer.OnBlockAddition(block)
	}
}
