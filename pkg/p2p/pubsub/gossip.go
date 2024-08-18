package pubsub

import (
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"chainnet/pkg/observer"
	"context"
	"fmt"

	pubSubP2P "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

const (
	// todo(): BlackListedNodes?
	TxMempoolPubSubTopic  = "txMempoolTopic"
	BlockAddedPubSubTopic = "blockAddedTopic"
)

var topicHandlers = map[string]func(ctx context.Context, sub *pubSubP2P.Subscription, netSubject observer.NetSubject){ //nolint:gochecknoglobals // this can be global var
	TxMempoolPubSubTopic:  listenForTxMempool,
	BlockAddedPubSubTopic: listenForBlocksAdded,
}

type Gossip struct {
	ctx    context.Context
	pubsub *pubSubP2P.PubSub

	encoder encoding.Encoding

	netSubject observer.NetSubject
	topicStore map[string]*pubSubP2P.Topic
}

func NewGossipPubSub(ctx context.Context, host host.Host, encoder encoding.Encoding, netSubject observer.NetSubject, topics []string, enableSubscribe bool) (*Gossip, error) {
	pubsub, err := pubSubP2P.NewGossipSub(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub module: %w", err)
	}

	topicStore := make(map[string]*pubSubP2P.Topic)
	// join the topics and subscribe/initialize handler if required
	for _, topicName := range topics {
		topic, errJoin := pubsub.Join(topicName)
		if errJoin != nil {
			return nil, fmt.Errorf("error joining pubsub topic %s: %w", topicName, errJoin)
		}

		// if subscribe is enabled, subscribe to the topic and initialize the handler. Otherwise, just join the
		// topic. Subscribe is not enabled for the cases in which we only want to publish to the topic (like wallets)
		// but not listen
		if enableSubscribe {
			// subscribe to the topic to listen
			sub, errSub := topic.Subscribe()
			if errSub != nil {
				return nil, fmt.Errorf("error subscribing to pubsub topic %s: %w", topicName, errSub)
			}

			// initialize handler
			if handler, ok := topicHandlers[topicName]; !ok {
				return nil, fmt.Errorf("unable to initialize handler for topic %s", topicName)
			} else if ok {
				go handler(ctx, sub, netSubject)
			}
		}

		// save the topics
		topicStore[topicName] = topic
	}

	return &Gossip{
		ctx:        ctx,
		pubsub:     pubsub,
		encoder:    encoder,
		netSubject: netSubject,
		topicStore: topicStore,
	}, nil
}

func listenForBlocksAdded(ctx context.Context, sub *pubSubP2P.Subscription, _ observer.NetSubject) {
	for {
		_, err := sub.Next(ctx)
		if err != nil {
			return
		}
	}

	//
}

func listenForTxMempool(ctx context.Context, sub *pubSubP2P.Subscription, netSubject observer.NetSubject) {
	for {
		_, err := sub.Next(ctx)
		if err != nil {
			return
		}

		netSubject.NotifyUnconfirmedTxReceived(kernel.Transaction{})
	}
}

func (g *Gossip) NotifyBlockAdded(_ kernel.Block) error {
	return nil
}

func (g *Gossip) SendTransaction(ctx context.Context, tx kernel.Transaction) error {
	topic, ok := g.topicStore[TxMempoolPubSubTopic]
	if !ok {
		return fmt.Errorf("topic %s not registered", TxMempoolPubSubTopic)
	}

	data, err := g.encoder.SerializeTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to serialize transaction: %w", err)
	}

	return topic.Publish(ctx, data)
}
