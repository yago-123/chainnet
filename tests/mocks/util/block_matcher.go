package util

import (
	"chainnet/pkg/kernel"
	"reflect"

	"github.com/stretchr/testify/mock"
)

// MatchByPreviousBlockPointer creates an argument matcher
// based on the PrevBlockHash field of a kernel.Block instance.
func MatchByPreviousBlockPointer(prevBlockHash []byte) interface{} {
	return mock.MatchedBy(func(innerBlock *kernel.Block) bool {
		return reflect.DeepEqual(innerBlock.PrevBlockHash, prevBlockHash)
	})
}

// MatchByPreviousBlock creates an argument matcher
// based on the PrevBlockHash field of a kernel.Block instance.
func MatchByPreviousBlock(prevBlockHash []byte) interface{} {
	return mock.MatchedBy(func(innerBlock kernel.Block) bool {
		return reflect.DeepEqual(innerBlock.PrevBlockHash, prevBlockHash)
	})
}
