package main

import (
	"fmt"
	"strings"
)

type blockchain struct {
	chain            []*block
	uncles           map[int]*block // Contains all current unreferenced uncles within reward range.
	referencedUncles []*block       // Contains uncles that has been referenced and has cashed in their reward
}

// Helpfull when in knowing if a block is dated.
func (bc *blockchain) round() int {
	return len(bc.chain)
}

func (bc *blockchain) addNewBlock(b *block) {
	//if !bytes.Equal(b.parentHash, bc.CurrentBlock().hash) {
	//panic("ABOUT TO ADD BLOCK WITH WRONG PARENT!")
	//}
	bc.chain = append(bc.chain, b)

	// Checks if any uncles is referencing other uncles. If they do -> Make block in main chain refer uncle instead
	for _, block := range bc.referencedUncles {
		if block.uncleBlocks != nil {
			bc.chain[block.depth].uncleBlocks = block.uncleBlocks
			bc.chain[block.depth].calcRewards()
			block.uncleBlocks = nil
			block.dat.rewardTot = block.dat.rewardTot - block.dat.rewardNephew
			block.dat.rewardNephew = 0
		}
	}
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
