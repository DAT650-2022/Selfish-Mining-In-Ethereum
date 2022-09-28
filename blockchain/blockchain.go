package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrTxNotFound    = errors.New("transaction not found")
	ErrNoValidTx     = errors.New("there is no valid transaction")
	ErrBlockNotFound = errors.New("block not found")
	ErrInvalidBlock  = errors.New("block is not valid")
)

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
	blocks []*Block
}

// NewBlockchain creates a new blockchain with genesis Block
func NewBlockchain(address string) (*Blockchain, error) {
	var blocks []*Block
	coinbaseTx, err := NewCoinbaseTX(address, GenesisCoinbaseData)
	if err != nil {
		return nil, err
	}
	genesis := NewGenesisBlock(TestBlockTime, coinbaseTx)
	blocks = append(blocks, genesis)
	return &Blockchain{blocks: blocks}, nil
}

// addBlock saves the block into the blockchain
func (bc *Blockchain) addBlock(block *Block) error {
	prevBlockHash := bc.CurrentBlock().Hash
	newBlock := NewBlock(block.Timestamp, block.Transactions, prevBlockHash)
	if bc.ValidateBlock(block) {
		bc.blocks = append(bc.blocks, newBlock)
		return nil
	}
	return ErrInvalidBlock
}

// GetGenesisBlock returns the Genesis Block
func (bc Blockchain) GetGenesisBlock() *Block {
	return bc.blocks[0]
}

// CurrentBlock returns the last block
func (bc Blockchain) CurrentBlock() *Block {
	return bc.blocks[len(bc.blocks)-1]
}

// GetBlock returns the block of a given hash
func (bc Blockchain) GetBlock(hash []byte) (*Block, error) {
	for _, i := range bc.blocks {
		if bytes.Equal(i.Hash, hash) {
			return i, nil
		}
	}
	return nil, ErrBlockNotFound
}

// ValidateBlock validates the block before adding it to the blockchain
func (bc *Blockchain) ValidateBlock(block *Block) bool {
	if block == nil {
		return false
	}
	if len(block.Transactions) < 1 {
		return false
	}
	pow := NewProofOfWork(block)
	return pow.Validate()
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) (*Block, error) {
	var verifiedTx []*Transaction
	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) {
			verifiedTx = append(verifiedTx, tx)
		}
	}

	if verifiedTx != nil {
		block := Block{Timestamp: TestBlockTime, Transactions: verifiedTx, PrevBlockHash: bc.CurrentBlock().Hash}
		block.Mine()
		bc.blocks = append(bc.blocks, &block)
		return bc.CurrentBlock(), nil
	}
	return nil, ErrNoValidTx
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTx, _ := bc.GetInputTXsOf(tx)
	return tx.Verify(prevTx)
}

// FindTransaction finds a transaction by its ID in the whole blockchain
func (bc Blockchain) FindTransaction(ID []byte) (*Transaction, error) {
	for _, block := range bc.blocks {
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return tx, nil
			}
		}
	}
	return nil, ErrTxNotFound
}

// FindUTXOSet finds and returns all unspent transaction outputs
func (bc Blockchain) FindUTXOSet() UTXOSet {
	utxo := make(UTXOSet)
	for _, block := range bc.blocks {
		utxo.Update(block.Transactions)
	}
	return utxo
}

// GetInputTXsOf returns a map index by the ID,
// of all transactions used as inputs in the given transaction
func (bc *Blockchain) GetInputTXsOf(tx *Transaction) (map[string]*Transaction, error) {
	prevTX := make(map[string]*Transaction)
	for _, vin := range tx.Vin {
		oldTx, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			return nil, err
		}
		prevTX[hex.EncodeToString(oldTx.ID)] = oldTx
	}
	return prevTX, nil
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) error {
	prevTx, err := bc.GetInputTXsOf(tx)
	if err != nil {
		return err
	}
	err = tx.Sign(privKey, prevTx)
	if err != nil {
		return err
	}
	return nil
}

func (bc Blockchain) String() string {
	var lines []string
	for _, block := range bc.blocks {
		lines = append(lines, fmt.Sprintf("%v", block))
	}
	return strings.Join(lines, "\n")
}
