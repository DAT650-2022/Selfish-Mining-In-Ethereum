package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"
)

// Block keeps block information
type Block struct {
	Timestamp     int64          // the block creation timestamp
	Transactions  []*Transaction // The block transactions
	PrevBlockHash []byte         // the hash of the previous block
	Hash          []byte         // the hash of the block
	Nonce         int            // the nonce of the block
}

// NewBlock creates and returns a non-mined Block
func NewBlock(timestamp int64, transactions []*Transaction, prevBlockHash []byte) *Block {
	return &Block{timestamp, transactions, prevBlockHash, nil, 0}
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(timestamp int64, tx *Transaction) *Block {
	return NewBlock(timestamp, []*Transaction{tx}, nil)
}

// Mine calculates and sets the block hash and nonce.
func (b *Block) Mine() {
	pow := NewProofOfWork(b)
	nonce, hash := pow.Run()
	if nonce == 0 {
		log.Fatalf("max nonce reached")
	}
	b.Nonce = nonce
	b.Hash = hash
}

// HashTransactions returns a hash of the transactions in the block
// This function iterates over all transactions in a block, serialize them
// and make a merkle tree of it.
// It return the merkle root hash.
func (b *Block) HashTransactions() []byte {
	var dataList [][]byte
	for _, tx := range b.Transactions {
		data := tx.Serialize()
		dataList = append(dataList, data)
	}
	merkleTree := NewMerkleTree(dataList)
	return merkleTree.MerkleRootHash()
}

// FindTransaction finds a transaction by its ID
func (b *Block) FindTransaction(ID []byte) (*Transaction, error) {
	for _, tx := range b.Transactions {
		if bytes.Equal(tx.ID, ID) {
			return tx, nil
		}
	}
	return nil, ErrTxNotFound
}

func (b *Block) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("============ Block %x ============", b.Hash))
	lines = append(lines, fmt.Sprintf("Prev. hash: %x", b.PrevBlockHash))
	lines = append(lines, fmt.Sprintf("Timestamp: %v", time.Unix(b.Timestamp, 0)))
	lines = append(lines, fmt.Sprintf("Nonce: %d", b.Nonce))
	lines = append(lines, fmt.Sprintf("Transactions:"))
	for i, tx := range b.Transactions {
		lines = append(lines, fmt.Sprintf("%d: %x", i, tx.ID))
	}
	return strings.Join(lines, "\n")
}
