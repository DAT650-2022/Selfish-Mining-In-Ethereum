package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// MerkleTree represents a Merkle tree
type MerkleTree struct {
	RootNode *Node
	Leafs    []*Node
}

// Node represents a Merkle tree node
type Node struct {
	Parent *Node
	Left   *Node
	Right  *Node
	Hash   []byte
}

const (
	leftNode = iota
	rightNode
)

// MerkleProof represents way to prove element inclusion on the merkle tree
type MerkleProof struct {
	proof [][]byte
	index []int64
}

// NewMerkleTree creates a new Merkle tree from a sequence of data
func NewMerkleTree(data [][]byte) *MerkleTree {
	if len(data) == 0 {
		panic("No merkle tree nodes")
	}

	var nodes []*Node
	// Special case: Single node
	if len(data) == 1 {
		merkle := NewMerkleNode(nil, nil, data[0])
		nodes = append(nodes, merkle)
		return &MerkleTree{RootNode: merkle, Leafs: nodes}
	}

	// Create leafs
	for _, i := range data {
		node := NewMerkleNode(nil, nil, i)
		nodes = append(nodes, node)
	}
	leafs := nodes

	// If we have an odd number of nodes -> Make copy of last node
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	// Build the layers of the Merkle Tree
	for i := 0; i < len(data)/2; i++ {
		// If only the root node exist in nodes -> Break
		if len(nodes) == 1 {
			break
		}

		// If we have an odd number of nodes -> Make copy of last node
		if len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}

		var newLayer []*Node
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(nodes[j], nodes[j+1], nil)
			newLayer = append(newLayer, node)
		}
		nodes = newLayer
	}
	return &MerkleTree{RootNode: nodes[0], Leafs: leafs}
}

// NewMerkleNode creates a new Merkle tree node
func NewMerkleNode(left, right *Node, data []byte) *Node {
	var hash [32]byte
	if left == nil && right == nil {
		hash = sha256.Sum256(data)
	} else {
		prevHash := append(left.Hash, right.Hash...)
		hash = sha256.Sum256(prevHash)
	}
	node := &Node{Left: left, Right: right, Hash: hash[:]}
	if left != nil && right != nil {
		left.Parent = node
		right.Parent = node
	}
	return node
}

// MerkleRootHash return the hash of the merkle root
func (mt *MerkleTree) MerkleRootHash() []byte {
	return mt.RootNode.Hash
}

// MakeMerkleProof returns a list of hashes and indexes required to
// reconstruct the merkle path of a given hash
//
// @param hash represents the hashed data (e.g. transaction ID) stored on
// the leaf node
// @return the merkle proof (list of intermediate hashes), a list of indexes
// indicating the node location in relation with its parent (using the
// constants: leftNode or rightNode), and a possible error.
func (mt *MerkleTree) MakeMerkleProof(hash []byte) ([][]byte, []int64, error) {
	// Find the leaf node containing the hash
	var currentNode *Node
	for _, i := range mt.Leafs {
		if bytes.Equal(i.Hash, hash) {
			currentNode = i
			break
		}
	}

	// If node doesnt exits
	if currentNode == nil {
		return [][]byte{}, []int64{}, fmt.Errorf("Node %x not found", hash)
	}

	// Find intermediate hashes and list of indexes
	hashes := [][]byte{}
	indexes := []int64{}
	for currentNode.Parent != nil {
		// Find out if current node is left or right of parent node
		if bytes.Equal(currentNode.Hash, currentNode.Parent.Left.Hash) { // Current node is to the left of the parent
			hashes = append(hashes, currentNode.Parent.Right.Hash) // Intermediate node is to the right of the parent
			indexes = append(indexes, rightNode)
		} else { //Current node is to the right of the parent
			hashes = append(hashes, currentNode.Parent.Left.Hash) // Intermediate node is to the left of the parent
			indexes = append(indexes, leftNode)
		}
		currentNode = currentNode.Parent
	}
	return hashes, indexes, nil
}

// VerifyProof verifies that the correct root hash can be retrieved by
// recreating the merkle path for the given hash and merkle proof.
//
// @param rootHash is the hash of the current root of the merkle tree
// @param hash represents the hash of the data (e.g. transaction ID)
// to be verified
// @param mProof is the merkle proof that contains the list of intermediate
// hashes and their location on the tree required to reconstruct
// the merkle path.
func VerifyProof(rootHash []byte, hash []byte, mProof MerkleProof) bool {
	currentHash := hash
	for index, i := range mProof.proof {
		var prevHash []byte
		if mProof.index[index] == rightNode { // If parent is to the right of the current node
			prevHash = append(currentHash, i...)
		} else { // If parent is to the left of the current node
			prevHash = append(i, currentHash...)
		}
		newHash := sha256.Sum256(prevHash)
		currentHash = newHash[:]
	}
	if bytes.Equal(currentHash, rootHash) {
		return true
	}
	return false
}
