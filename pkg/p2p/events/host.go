package events

import (
	"chainnet/pkg/observer"
	"context"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/sirupsen/logrus"
)

// todo(): turn this into a struct with administration methods if we rely more on this event bus system from host
// InitializeHostEventsSubscription creates the subscription and the listener for host events
func InitializeHostEventsSubscription(ctx context.Context, logger *logrus.Logger, host host.Host, subject observer.NetSubject) error {
	sub, err := host.EventBus().Subscribe(new(event.EvtPeerIdentificationCompleted))
	if err != nil {
		return err
	}

	go listenForHostEvents(ctx, logger, sub, subject)

	return nil
}

// listenForHostEvents represents the event listener that reacts to events emitter by the host event bus
func listenForHostEvents(ctx context.Context, logger *logrus.Logger, sub event.Subscription, subject observer.NetSubject) {
	for {
		select {
		case evt := <-sub.Out():
			switch e := evt.(type) {
			case event.EvtPeerIdentificationCompleted:
				subject.NotifyNodeDiscovered(e.Peer)
				break
			default:
				logger.Errorf("unhandled event type: %T", evt)
			}
		case <-ctx.Done():
			// context finished, stop listening for events
			logger.Errorf("context canceled, stopping event listener")
			return
		}
	}
}
