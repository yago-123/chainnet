package bloom

import (
	"chainnet/pkg/kernel"
	"encoding/binary"
	"fmt"
)

const (
	BloomObserverID = "bloomfilter"
	BloomFilterSize = 32
)

type BlockBloomFilter struct {
	bloom map[string][]bool
}

func NewBlockBloomFilter() *BlockBloomFilter {
	return &BlockBloomFilter{
		bloom: make(map[string][]bool),
	}
}

func (bf *BlockBloomFilter) AddBlock(block *kernel.Block) {
	bf.bloom[string(block.Hash)] = make([]bool, BloomFilterSize)

	for _, tx := range block.Transactions {
		index := binary.BigEndian.Uint64(tx.ID) % BloomFilterSize
		bf.bloom[string(block.Hash)][index] = true
	}
}

func (bf *BlockBloomFilter) PresentInBlock(txID, blockID []byte) (bool, error) {
	if filter, ok := bf.bloom[string(blockID)]; ok {
		index := binary.BigEndian.Uint64(txID) % BloomFilterSize
		return filter[index], nil
	}

	return false, fmt.Errorf("block %s not found in bloom filter", string(blockID))
}

func (bf *BlockBloomFilter) ID() string {
	return BloomObserverID
}

func (bf *BlockBloomFilter) OnBlockAddition(block *kernel.Block) {
	// calculate the bloom filter of the new block
	bf.AddBlock(block)
}
