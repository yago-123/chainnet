package util

import (
	"chainnet/pkg/block"
	"github.com/stretchr/testify/mock"
	"reflect"
)

// MatchByPreviousBlockPointer creates an argument matcher
// based on the PrevBlockHash field of a block.Block instance.
func MatchByPreviousBlockPointer(prevBlockHash []byte) interface{} {
	return mock.MatchedBy(func(innerBlock *block.Block) bool {
		return reflect.DeepEqual(innerBlock.PrevBlockHash, prevBlockHash)
	})
}

// MatchByPreviousBlock creates an argument matcher
// based on the PrevBlockHash field of a block.Block instance.
func MatchByPreviousBlock(prevBlockHash []byte) interface{} {
	return mock.MatchedBy(func(innerBlock block.Block) bool {
		return reflect.DeepEqual(innerBlock.PrevBlockHash, prevBlockHash)
	})
}
