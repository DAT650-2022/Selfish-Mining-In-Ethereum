package main

type block struct {
	hash          []byte
	prevBlock     *block
	prevBlockHash []byte
	uncleBlocks   []*block
	difficulty    int
	dat           dataUnit
}

// TODO: maybe move selfish bool to block?
// may not matter?
type dataUnit struct {
	rewardTot    int
	rewardMined  int
	rewardNephew int
	rewardUncle  int
	selfish      bool
}

func newGenesisBlock() *block {
	return &block{
		hash:          []byte("Darwin"),
		prevBlock:     nil,
		prevBlockHash: nil,
		uncleBlocks:   nil,
		difficulty:    0,
		dat:           dataUnit{rewardTot: 0},
	}
}
