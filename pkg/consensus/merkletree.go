package consensus

import (
	"bytes"
	"chainnet/pkg/consensus/util"
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

// newMerkleNode creates a new Merkle node
func newMerkleNode(left, right *MerkleNode, data []byte, hasher hash.Hashing) (*MerkleNode, error) {
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
			return nil, fmt.Errorf("error hashing left (%s) and right (%s) nodes: %w", leftHash, rightHash, err)
		}
		node.Hash = nodeHash
	}

	node.Left = left
	node.Right = right

	return &node, nil
}

// MerkleTree represents a Merkle tree
type MerkleTree struct {
	root *MerkleNode
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

			parent, err := newMerkleNode(&left, &right, nil, hasher)
			if err != nil {
				return nil, fmt.Errorf("error creating Merkle parent node for left (%s) and right (%s) nodes: %w", left.Hash, right.Hash, err)
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
		node, err := newMerkleNode(nil, nil, proof, hasher)
		if err != nil {
			return nil, fmt.Errorf("error creating Merkle node for hash %s: %w", proof, err)
		}

		nodes = append(nodes, *node)
	}

	return newMerkleTreeFromNodes(nodes, hasher)
}

// NewMerkleTreeFromTxs creates a new Merkle tree from a list of transactions
func NewMerkleTreeFromTxs(txs []*kernel.Transaction, hasher hash.Hashing) (*MerkleTree, error) {
	var nodes []MerkleNode

	// create leaf nodes using the Assemble method
	for _, tx := range txs {
		// make sure that the hash of the transaction is correct
		txHash, err := util.CalculateTxHash(tx, hasher)
		if err != nil {
			return nil, fmt.Errorf("error calculating hash for transaction %s: %w", tx.ID, err)
		}

		// verify computed hash natch the predefined hash
		if !bytes.Equal(txHash, tx.ID) {
			return nil, fmt.Errorf("transaction %s has invalid hash, expected %s", tx.ID, txHash)
		}

		// generate new node
		node, err := newMerkleNode(nil, nil, tx.ID, hasher)
		if err != nil {
			return nil, fmt.Errorf("error creating Merkle node for transaction %s: %w", tx.ID, err)
		}
		nodes = append(nodes, *node)
	}

	return newMerkleTreeFromNodes(nodes, hasher)
}

// RootHash returns the root hash of the Merkle tree
func (mt *MerkleTree) RootHash() []byte {
	return mt.root.Hash
}
