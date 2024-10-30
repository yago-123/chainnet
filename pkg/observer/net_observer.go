package observer

import (
	"sync"

	"github.com/yago-123/chainnet/pkg/kernel"

	"github.com/libp2p/go-libp2p/core/peer"
)

// NetObserver interface that defines the methods that a network observer should implement
type NetObserver interface {
	ID() string
	OnNodeDiscovered(peerID peer.ID)
	OnUnconfirmedHeaderReceived(peer peer.ID, header kernel.BlockHeader)
	OnUnconfirmedTxReceived(peer peer.ID, tx kernel.Transaction)
	OnUnconfirmedTxIDReceived(peer peer.ID, txID string)
}

// NetSubject controller that manages the net observers
type NetSubject interface {
	Register(observer NetObserver)
	Unregister(observer NetObserver)
	NotifyNodeDiscovered(peerID peer.ID)
	NotifyUnconfirmedHeaderReceived(peer peer.ID, header kernel.BlockHeader)
	NotifyUnconfirmedTxReceived(peer peer.ID, tx kernel.Transaction)
	NotifyUnconfirmedTxIDReceived(peer peer.ID, txID string)
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
func (no *NetSubjectController) NotifyNodeDiscovered(peerID peer.ID) {
	no.mu.Lock()
	defer no.mu.Unlock()
	for _, observer := range no.observers {
		observer.OnNodeDiscovered(peerID)
	}
}

// NotifyUnconfirmedHeaderReceived notifies all observers that a new block has been added
func (no *NetSubjectController) NotifyUnconfirmedHeaderReceived(peer peer.ID, header kernel.BlockHeader) {
	no.mu.Lock()
	defer no.mu.Unlock()
	for _, observer := range no.observers {
		observer.OnUnconfirmedHeaderReceived(peer, header)
	}
}

// NotifyUnconfirmedTxReceived notifies all observers that a new unconfirmed transaction has been received
func (no *NetSubjectController) NotifyUnconfirmedTxReceived(peer peer.ID, tx kernel.Transaction) {
	no.mu.Lock()
	defer no.mu.Unlock()
	for _, observer := range no.observers {
		observer.OnUnconfirmedTxReceived(peer, tx)
	}
}

// NotifyUnconfirmedTxIDReceived notifies all observers that a new unconfirmed transaction ID has been received
func (no *NetSubjectController) NotifyUnconfirmedTxIDReceived(peer peer.ID, txID string) {
	no.mu.Lock()
	defer no.mu.Unlock()
	for _, observer := range no.observers {
		observer.OnUnconfirmedTxIDReceived(peer, txID)
	}
}
