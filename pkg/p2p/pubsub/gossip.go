package pubsub

import (
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/kernel"
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

	netSubject observer.NetSubject
	topicStore map[string]*pubSubP2P.Topic
}

func NewGossipPubSub(ctx context.Context, host host.Host, netSubject observer.NetSubject, topics []string) (*Gossip, error) {
	pubsub, errGossip := pubSubP2P.NewGossipSub(ctx, host)
	if errGossip != nil {
		return nil, fmt.Errorf("failed to create pubsub module: %w", errGossip)
	}

	topicStore := make(map[string]*pubSubP2P.Topic)
	// join the topics and subscribe
	for _, topicName := range topics {
		topic, err := pubsub.Join(topicName)
		if err != nil {
			return nil, fmt.Errorf("error joining pubsub topic %s: %w", topicName, err)
		}

		// subscribe to the topic to listen
		sub, err := topic.Subscribe()
		if err != nil {
			return nil, fmt.Errorf("error subscribing to pubsub topic %s: %w", topicName, err)
		}

		// initialize handler
		if handler, ok := topicHandlers[topicName]; !ok {
			return nil, fmt.Errorf("unable to initialize handler for topic %s", topicName)
		} else if ok {
			go handler(ctx, sub, netSubject)
		}

		// save the topics
		topicStore[topicName] = topic
	}

	return &Gossip{
		ctx:        ctx,
		pubsub:     pubsub,
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
