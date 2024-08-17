package pubsub

import (
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

type Gossip struct {
	ctx    context.Context
	pubsub *pubSubP2P.PubSub

	topics map[string]*pubSubP2P.Topic
}

func NewGossipPubSub(ctx context.Context, host host.Host) (*Gossip, error) {
	pubsub, err := pubSubP2P.NewGossipSub(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	// joing the block added topic
	topic, err := pubsub.Join(BlockAddedPubSubTopic)
	if err != nil {
		return nil, fmt.Errorf("error joining pubsub topic: %w", err)
	}

	// subscribe to the topic to listen for new blocks
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("error subscribing to pubsub topic: %w", err)
	}

	// listen messages from the topic
	go listenForBlocksAdded(ctx, sub)

	// save the topics
	topics := make(map[string]*pubSubP2P.Topic)
	topics[BlockAddedPubSubTopic] = topic

	return &Gossip{
		ctx:    ctx,
		pubsub: pubsub,
		topics: topics,
	}, nil
}

func listenForBlocksAdded(ctx context.Context, sub *pubSubP2P.Subscription) {
	for {
		_, err := sub.Next(ctx)
		if err != nil {
			return
		}

		// fmt.Println("Received message:", string(msg.Data))
	}
}

func (g *Gossip) NotifyBlockAdded(_ kernel.Block) error {
	return nil
}
