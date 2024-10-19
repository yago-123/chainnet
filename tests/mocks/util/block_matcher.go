package util

import (
	"reflect"

	"github.com/yago-123/chainnet/pkg/kernel"

	"github.com/stretchr/testify/mock"
)

// MatchByPreviousBlockPointer creates an argument matcher
// based on the PrevBlockHash field of a kernel.Block instance.
func MatchByPreviousBlockPointer(prevBlockHash []byte) interface{} {
	return mock.MatchedBy(func(innerBlock *kernel.Block) bool {
		return reflect.DeepEqual(innerBlock.Header.PrevBlockHash, prevBlockHash)
	})
}

// MatchByPreviousBlock creates an argument matcher
// based on the PrevBlockHash field of a kernel.Block instance.
func MatchByPreviousBlock(prevBlockHash []byte) interface{} {
	return mock.MatchedBy(func(innerBlock kernel.Block) bool {
		return reflect.DeepEqual(innerBlock.Header.PrevBlockHash, prevBlockHash)
	})
}
