package pubsub

import (
	"chainnet/pkg/kernel"
	"context"
)

type PubSub interface {
	NotifyBlockHeaderAdded(ctx context.Context, header kernel.BlockHeader) error
	NotifyTransactionAdded(ctx context.Context, tx kernel.Transaction) error
}
