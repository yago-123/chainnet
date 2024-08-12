package util

import (
	"context"
)

type CtxMutex struct {
	ch chan struct{}
}

func (mu *CtxMutex) Lock(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case mu.ch <- struct{}{}:
		return true
	}
}

func (mu *CtxMutex) Unlock() {
	<-mu.ch
}

func (mu *CtxMutex) Locked() bool {
	return len(mu.ch) > 0
}
