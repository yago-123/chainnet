package pubsub

import (
	"chainnet/config"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"chainnet/pkg/observer"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	pubSubP2P "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

const (
	// todo(): BlackListedNodes?
	TxMempoolPubSubTopic  = "tx-mempool-topic"
	BlockAddedPubSubTopic = "block-added-topic"
)

type gossipHandler struct {
	ctx        context.Context
	logger     *logrus.Logger
	encoder    encoding.Encoding
	netSubject observer.NetSubject
}

func newGossipHandler(ctx context.Context, logger *logrus.Logger, encoder encoding.Encoding, netSubject observer.NetSubject) *gossipHandler {
	return &gossipHandler{
		ctx:        ctx,
		logger:     logger,
		encoder:    encoder,
		netSubject: netSubject,
	}
}

// listenForBlocksAdded represents the handler for the block added topic
func (h *gossipHandler) listenForBlocksAdded(sub *pubSubP2P.Subscription) {
	for {
		msg, err := sub.Next(h.ctx)
		if err != nil {
			return
		}

		// todo(): should we send the block or the block header?

		block, err := h.encoder.DeserializeBlock([]byte(msg.String()))
		if err != nil {
			h.logger.Errorf("failed deserializing block from %s: %s", msg.ReceivedFrom, err)
		}

		h.logger.Debugf("received block from %s with block ID %x", msg.ReceivedFrom, block.Hash)

		h.netSubject.NotifyUnconfirmedBlockReceived(*block)
	}
}

// listenForTxMempool represents the handler for the tx mempool topic
func (h *gossipHandler) listenForTxMempool(sub *pubSubP2P.Subscription) {
	for {
		_, err := sub.Next(h.ctx)
		if err != nil {
			return
		}

		// tx, err := h.encoder.DeserializeTransaction([]byte(msg.String()))
		// if err != nil {
		//	h.logger.Errorf("failed deserializing transaction: %s", err)
		// }

		// h.netSubject.NotifyUnconfirmedTxReceived(*tx)
	}
}

type GossipPubSub struct {
	ctx    context.Context
	pubsub *pubSubP2P.PubSub

	encoder encoding.Encoding

	netSubject observer.NetSubject
	topicStore map[string]*pubSubP2P.Topic
}

func NewGossipPubSub(ctx context.Context, cfg *config.Config, host host.Host, encoder encoding.Encoding, netSubject observer.NetSubject, topics []string, enableSubscribe bool) (*GossipPubSub, error) {
	pubsub, err := pubSubP2P.NewGossipSub(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub module: %w", err)
	}

	handler := newGossipHandler(ctx, cfg.Logger, encoder, netSubject)

	// initialize handlers for the topics available
	topicHandlers := map[string]func(sub *pubSubP2P.Subscription){
		TxMempoolPubSubTopic:  handler.listenForTxMempool,
		BlockAddedPubSubTopic: handler.listenForBlocksAdded,
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
		// todo: put enableSubscribe as flag
		if enableSubscribe {
			// subscribe to the topic to listen
			sub, errSub := topic.Subscribe()
			if errSub != nil {
				return nil, fmt.Errorf("error subscribing to pubsub topic %s: %w", topicName, errSub)
			}

			// start handlers
			if handlerFunc, ok := topicHandlers[topicName]; !ok {
				return nil, fmt.Errorf("unable to initialize handler for topic %s", topicName)
			} else if ok {
				go handlerFunc(sub)
			}
		}

		// save the topics
		topicStore[topicName] = topic
	}

	return &GossipPubSub{
		ctx:        ctx,
		pubsub:     pubsub,
		encoder:    encoder,
		netSubject: netSubject,
		topicStore: topicStore,
	}, nil
}

// NotifyBlockAdded used for notifying the pubsub network that a local block has been added to the blockchain
func (g *GossipPubSub) NotifyBlockAdded(_ kernel.Block) error {
	// todo(): should we control which blocks are sent to the pub sub net? (e.g. only blocks that are mined locally?)
	return nil
}

func (g *GossipPubSub) NotifyTransactionAdded(ctx context.Context, tx kernel.Transaction) error {
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
