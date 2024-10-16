package pubsub

import (
	"chainnet/pkg/kernel"
	"context"
)

type PubSub interface {
	NotifyBlockAdded(ctx context.Context, block kernel.Block) error
	NotifyTransactionAdded(ctx context.Context, tx kernel.Transaction) error
}
