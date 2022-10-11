package main

type block struct {
	hash        []byte
	parent      *block
	parentHash  []byte
	uncleBlocks []*block
	difficulty  int
	dat         dataUnit
	depth       int
}

// TODO: maybe move selfish bool to block?
// may not matter?
type dataUnit struct {
	rewardTot    float64
	rewardMined  int
	rewardNephew float64
	rewardUncle  float64
	selfish      bool
}

func newGenesisBlock() *block {
	return &block{
		hash:        []byte("Darwin"),
		parent:      nil,
		parentHash:  nil,
		uncleBlocks: nil,
		difficulty:  0,
		dat:         dataUnit{rewardTot: 0},
		depth:       0,
	}
}

func (b *block) calcRewards() {
	// b.dat.rewardUncle needs to be calculated
	// later since it can't know ahead of time if its included
	// as uncle in future blocks
	b.dat.rewardMined = BLOCKREWARD
	b.dat.rewardUncle = 0
	nephew := (1.00 / 36.00) * float64(len(b.uncleBlocks)) * float64(BLOCKREWARD)
	b.dat.rewardNephew = nephew
	b.calcTotal()
}

func (b *block) updateUncle(reward float64) {
	b.dat.rewardMined = 0
	b.dat.rewardNephew = 0
	b.dat.rewardUncle = reward
	b.calcTotal() // Update total
}

func (b *block) calcTotal() {
	b.dat.rewardTot = float64(b.dat.rewardMined) + b.dat.rewardNephew + b.dat.rewardUncle
}

//func (b *block) String() string {
//return fmt.Sprintf("Hash:\t%s \nPHash:\t%s \nDepth:\t%d \nReward:\t%f \nSelfish:\t%t", string(b.hash), string(b.parentHash), b.depth, b.dat.rewardTot, b.dat.selfish)
//}
