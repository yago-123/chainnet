package observer

import (
	"sync"

	"github.com/yago-123/chainnet/pkg/kernel"
)

// ChainObserver interface that defines the methods that a block observer should implement
type ChainObserver interface {
	ID() string
	OnBlockAddition(block *kernel.Block)
	OnTxAddition(tx *kernel.Transaction)
}

// ChainSubject controller that manages the block observers
type ChainSubject interface {
	Register(observer ChainObserver)
	Unregister(observer ChainObserver)
	NotifyBlockAdded(block *kernel.Block)
	NotifyTxAdded(tx *kernel.Transaction)
}

type ChainSubjectController struct {
	observers map[string]ChainObserver
	mu        sync.Mutex
}

func NewChainSubject() *ChainSubjectController {
	return &ChainSubjectController{
		observers: make(map[string]ChainObserver),
	}
}

// Register adds an observer to the list of observers
func (so *ChainSubjectController) Register(observer ChainObserver) {
	so.mu.Lock()
	defer so.mu.Unlock()
	so.observers[observer.ID()] = observer
}

// Unregister removes an observer from the list of observers
func (so *ChainSubjectController) Unregister(observer ChainObserver) {
	so.mu.Lock()
	defer so.mu.Unlock()
	delete(so.observers, observer.ID())
}

// NotifyBlockAdded notifies all observers that a new block has been added
func (so *ChainSubjectController) NotifyBlockAdded(block *kernel.Block) {
	so.mu.Lock()
	defer so.mu.Unlock()
	for _, observer := range so.observers {
		observer.OnBlockAddition(block)
	}
}

// NotifyTxAdded notifies all observers that a new transaction has been added
func (so *ChainSubjectController) NotifyTxAdded(tx *kernel.Transaction) {
	so.mu.Lock()
	defer so.mu.Unlock()
	for _, observer := range so.observers {
		observer.OnTxAddition(tx)
	}
}
