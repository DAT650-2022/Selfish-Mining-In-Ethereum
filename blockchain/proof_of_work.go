package main

import (
	"bytes"
	"crypto/sha256"
	"math"
	"math/big"
)

var maxNonce = math.MaxInt64

// TARGETBITS define the mining difficulty
const TARGETBITS = 8

// ProofOfWork represents a block mined with a target difficulty
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork builds a ProofOfWork
func NewProofOfWork(block *Block) *ProofOfWork {
	target := *big.NewInt(0)
	target.SetBit(&target, 256-TARGETBITS, 1)
	return &ProofOfWork{block: block, target: &target}
}

// setupHeader prepare the header of the block
func (pow *ProofOfWork) setupHeader() []byte {
	var header []byte
	header = append(header, pow.block.PrevBlockHash...)
	header = append(header, pow.block.HashTransactions()...)
	header = append(header, IntToHex(pow.block.Timestamp)...)
	header = append(header, IntToHex(TARGETBITS)...)
	return header
}

// addNonce adds a nonce to the header
func addNonce(nonce int, header []byte) []byte {
	header = append(header, IntToHex(int64(nonce))...)
	return header
}

// Run performs the proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
	nonce := 1
	header := pow.setupHeader()
	for {
		if nonce > maxNonce {
			break
		}
		nonceHeader := addNonce(nonce, header)
		hash := sha256.Sum256(nonceHeader)
		hashNumber := *big.NewInt(0)
		hashNumber.SetBytes(hash[:])
		if hashNumber.Cmp(pow.target) == -1 {
			return nonce, hash[:]
		}
		nonce += 1
	}
	return 0, nil
}

// Validate validates block's Proof-Of-Work
// This function just validates if the block header hash
// is less than the target AND equals to the mined block hash.
func (pow *ProofOfWork) Validate() bool {
	header := addNonce(pow.block.Nonce, pow.setupHeader())
	hashHeader := sha256.Sum256(header[:])
	hashNumber := *big.NewInt(0)
	hashNumber.SetBytes(hashHeader[:])
	if hashNumber.Cmp(pow.target) == -1 && bytes.Equal(hashHeader[:], pow.block.Hash) {
		return true
	}
	return false
}
