package main

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

func newBlockChain() *blockchain {

	return &blockchain{
		chain:  []*block{newGenesisBlock()},
		uncles: make(map[string]*block, 0),
	}
}
