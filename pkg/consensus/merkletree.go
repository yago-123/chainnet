package consensus

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

// NewMerkleNode creates a new Merkle node
func NewMerkleNode(left, right *MerkleNode, data []byte, hasher hash.Hashing) (*MerkleNode, error) {
	node := MerkleNode{}

	// in case is leaf node, assign hash directly
	if left == nil && right == nil {
		node.Hash = data
	}

	leftHash := []byte{}
	rightHash := []byte{}
	// in case is not leaf node, hash the left and right nodes
	if left != nil || right != nil {
		if left != nil {
			leftHash = left.Hash
		}

		if right != nil {
			rightHash = right.Hash
		}

		nodeHash, err := hasher.Hash(append(leftHash, rightHash...))
		if err != nil {
			return nil, fmt.Errorf("error hashing left (%s) and right (%s) nodes: %v", leftHash, rightHash, err)
		}
		node.Hash = nodeHash
	}

	node.Left = left
	node.Right = right

	return &node, nil
}

// MerkleTree represents a Merkle tree
type MerkleTree struct {
	Root *MerkleNode
}

// newMerkleTreeFromNodes creates a new Merkle tree from a list of Merkle nodes
func newMerkleTreeFromNodes(nodes []MerkleNode, hasher hash.Hashing) (*MerkleTree, error) {
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

			parent, err := NewMerkleNode(&left, &right, nil, hasher)
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

func NewMerkleTreeFromHashes(proofs [][]byte, hasher hash.Hashing) (*MerkleTree, error) {
	var nodes []MerkleNode

	if len(proofs) == 0 {
		return nil, fmt.Errorf("no proofs were provided")
	}

	for _, proof := range proofs {
		node, err := NewMerkleNode(nil, nil, proof, hasher)
		if err != nil {
			return nil, fmt.Errorf("error creating Merkle node for hash %s: %v", proof, err)
		}

		nodes = append(nodes, *node)
	}

	return newMerkleTreeFromNodes(nodes, hasher)
}

// NewMerkleTreeFromTxs creates a new Merkle tree from a list of transactions
func NewMerkleTreeFromTxs(txs []*kernel.Transaction, hasher hash.Hashing) (*MerkleTree, error) {
	var nodes []MerkleNode

	// create leaf nodes using the Assemble method
	for _, txn := range txs {
		node, err := NewMerkleNode(nil, nil, txn.ID, hasher)
		if err != nil {
			return nil, fmt.Errorf("error creating Merkle node for transaction %s: %v", txn.ID, err)
		}
		nodes = append(nodes, *node)
	}

	return newMerkleTreeFromNodes(nodes, hasher)
}
