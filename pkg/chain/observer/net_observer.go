package observer

import "sync"

// NetObserver interface that defines the methods that a network observer should implement
type NetObserver interface {
	ID() string
	OnNodeDiscovered(peerID string)
}

// NetSubject controller that manages the net observers
type NetSubject interface {
	Register(observer NetObserver)
	Unregister(observer NetObserver)
	NotifyNodeDiscovered(peerID string)
}

type NetSubjectController struct {
	observers map[string]NetObserver
	mu        sync.Mutex
}

func NewNetSubject() *NetSubjectController {
	return &NetSubjectController{
		observers: make(map[string]NetObserver),
	}
}

// Register adds an observer to the list of observers
func (no *NetSubjectController) Register(observer NetObserver) {
	no.mu.Lock()
	defer no.mu.Unlock()
	no.observers[observer.ID()] = observer
}

// Unregister removes an observer from the list of observers
func (no *NetSubjectController) Unregister(observer NetObserver) {
	no.mu.Lock()
	defer no.mu.Unlock()
	delete(no.observers, observer.ID())
}

// NotifyNodeDiscovered notifies all observers that a new node has been discovered
func (no *NetSubjectController) NotifyNodeDiscovered(peerID string) {
	no.mu.Lock()
	defer no.mu.Unlock()
	for _, observer := range no.observers {
		observer.OnNodeDiscovered(peerID)
	}
}