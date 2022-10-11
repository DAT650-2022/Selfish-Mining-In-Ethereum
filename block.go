package main

type block struct {
	hash        []byte
	parent      *block
	parentHash  []byte
	uncleBlocks []*block
	dat         dataUnit
	depth       int
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
		hash:        []byte("Darwin"),
		parent:      nil,
		parentHash:  nil,
		uncleBlocks: nil,
		dat:         dataUnit{rewardTot: 0},
		depth:       0,
	}
}
