package miner

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"fmt"
)

// MerkleNode represents a node in the Merkle tree
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Hash  []byte
}

// MerkleTree represents a Merkle tree
type MerkleTree struct {
	Root *MerkleNode
}

// NewMerkleNode creates a new Merkle node
func NewMerkleNode(left, right *MerkleNode, data []byte) (*MerkleNode, error) {
	node := MerkleNode{}
	hasher := hash.NewSHA256()

	if left == nil && right == nil {
		// this is a leaf node
		node.Hash = data
	}

	if left != nil || right != nil {
		// this is a non-leaf node
		nodeHash, err := hasher.Hash(append(left.Hash, right.Hash...))
		if err != nil {
			return nil, fmt.Errorf("error hashing left (%s) and right (%s) nodes: %v", left.Hash, right.Hash, err)
		}
		node.Hash = nodeHash
	}

	node.Left = left
	node.Right = right

	return &node, nil
}

// NewMerkleTree creates a new Merkle tree from a list of transactions
func NewMerkleTree(txs []*kernel.Transaction) (*MerkleTree, error) {
	var nodes []MerkleNode

	// create leaf nodes using the Assemble method
	for _, txn := range txs {
		node, err := NewMerkleNode(nil, nil, txn.ID)
		if err != nil {
			return nil, fmt.Errorf("error creating Merkle node for transaction %s: %v", txn.ID, err)
		}
		nodes = append(nodes, *node)
	}

	// create the Merkle tree
	for len(nodes) > 1 {
		var newLevel []MerkleNode

		for i := 0; i < len(nodes); i += 2 {
			var left, right MerkleNode
			left = nodes[i]

			if i+1 < len(nodes) {
				right = nodes[i+1]
			}

			if i+1 >= len(nodes) {
				right = nodes[i] // in case of odd number of nodes, duplicate the last node
			}

			parent, err := NewMerkleNode(&left, &right, nil)
			if err != nil {
				return nil, fmt.Errorf("error creating Merkle parent node for left (%s) and right (%s) nodes: %v", left.Hash, right.Hash, err)
			}
			newLevel = append(newLevel, *parent)
		}

		nodes = newLevel
	}

	tree := MerkleTree{&nodes[0]}

	return &tree, nil
}
