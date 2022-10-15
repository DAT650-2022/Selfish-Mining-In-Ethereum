package main

import (
	"bytes"
	"fmt"
	"strings"
)

type blockchain struct {
	chain            []*block
	uncles           map[int]*block // Contains all current unreferenced uncles within reward range.
	referencedUncles []*block       // Contains uncles that has been referenced and has cashed in their reward
	name             string
}

func (bc *blockchain) addNewBlock(b *block) {
	if !bytes.Equal(b.parentHash, bc.CurrentBlock().hash) {
		fmt.Println("ABOUT TO ADD BLOCK WITH WRONG PARENT!")
	}

	if len(bc.chain)-1%100 == 0 { // offset the genesis block from count
		println(fmt.Sprintf("adding block: %d, for chain %s", len(bc.chain), bc.name))
	}

	for _, block := range bc.referencedUncles {
		if block.uncleBlocks != nil {
			bc.chain[block.depth].uncleBlocks = block.uncleBlocks
			bc.chain[block.depth].calcRewards()
			block.uncleBlocks = nil
			block.dat.rewardTot = block.dat.rewardTot - block.dat.rewardNephew
			block.dat.rewardNephew = 0
		}
	}
	bc.chain = append(bc.chain, b)
}

func (bc *blockchain) CurrentBlock() *block {
	return bc.chain[len(bc.chain)-1]
}

func newBlockChain() *blockchain {
	return &blockchain{
		chain:            []*block{newGenesisBlock()},
		uncles:           make(map[int]*block, 0),
		referencedUncles: []*block{},
	}
}

func (bc *blockchain) String() string {
	var lines []string
	for _, block := range bc.chain {
		lines = append(lines, fmt.Sprintf("%v", block))
	}
	return strings.Join(lines, "\n")
}

func (bc *blockchain) StringUncles() string {
	var lines []string
	for _, block := range bc.referencedUncles {
		lines = append(lines, fmt.Sprintf("%v", block))
	}

	return strings.Join(lines, "\n")
}
