package events

import (
	"chainnet/pkg/observer"
	"context"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/sirupsen/logrus"
)

func InitializeHostEventsSubscription(ctx context.Context, logger *logrus.Logger, host host.Host, subject observer.NetSubject) error {
	sub, err := host.EventBus().Subscribe(new(event.EvtPeerIdentificationCompleted))
	if err != nil {
		return err
	}

	go listenForHostEvents(ctx, logger, sub, subject)

	return nil
}

func listenForHostEvents(ctx context.Context, logger *logrus.Logger, sub event.Subscription, subject observer.NetSubject) {
	for {
		select {
		case evt := <-sub.Out():
			switch e := evt.(type) {
			case event.EvtPeerIdentificationCompleted:
				subject.NotifyNodeDiscovered(e.Peer)
			default:
				// log error message for unknown event types
				logger.Errorf("unhandled event type: %T", evt)
			}
		case <-ctx.Done():
			// context finished, stop listening for events
			logger.Errorf("context canceled, stopping event listener")
			return
		}
	}
}
