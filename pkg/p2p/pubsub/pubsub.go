package pubsub

import (
	"chainnet/pkg/kernel"
	"context"
)

type PubSub interface {
	NotifyBlockAdded(block kernel.Block) error
	NotifyTransactionAdded(ctx context.Context, tx kernel.Transaction) error
}
