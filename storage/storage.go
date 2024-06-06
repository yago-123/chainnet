package storage

import "chainnet/block"

type Storage interface {
	NumberOfBlocks() (uint, error)
	PersistBlock(block block.Block) error
}
