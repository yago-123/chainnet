package pubsub

import "chainnet/pkg/kernel"

type PubSub interface {
	NotifyBlockAdded(block kernel.Block) error
}
