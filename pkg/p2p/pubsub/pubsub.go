package pubsub

import (
	"context"

	"github.com/yago-123/chainnet/pkg/kernel"
)

type PubSub interface {
	NotifyBlockHeaderAdded(ctx context.Context, header kernel.BlockHeader) error
	NotifyTransactionAdded(ctx context.Context, tx kernel.Transaction) error
}
