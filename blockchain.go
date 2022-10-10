package main

import (
	"fmt"
	"strings"
)

type blockchain struct {
	chain  []*block
	uncles map[string]*block
}

// Helpfull when in knowing if a block is dated.
func (bc *blockchain) round() int {
	return len(bc.chain)
}
func (bc *blockchain) addNewBlock(b *block) {
	bc.chain = append(bc.chain, b)
}

func (bc *blockchain) CurrentBlock() *block {
	return bc.chain[len(bc.chain)-1]
}

func newBlockChain() *blockchain {
	return &blockchain{
		chain:  []*block{newGenesisBlock()},
		uncles: make(map[string]*block, 0),
	}
}

func (bc *blockchain) String() string {
	var lines []string
	for _, block := range bc.chain {
		lines = append(lines, fmt.Sprintf("%v", block))
	}
	return strings.Join(lines, "\n")
}
