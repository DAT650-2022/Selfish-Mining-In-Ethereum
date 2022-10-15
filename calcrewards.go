package main

import "fmt"

type chainTots struct {
	total, totalSelf, totlaHonest float64
	uncleSelf, uncleHonest        float64
	nephewSelf, nephewHonest      float64
	minedSelf, minedHonest        float64

	absSelfRev, absHonestRev float64
}

func calcChainRewards(bc *blockchain) *chainTots {
	tots := chainTots{}
	for i := 1; i < len(bc.chain); i++ { //  starting at one to ignore genesis block
		block := bc.chain[i]
		uncleReward(block, &tots)
		if block.dat.selfish {
			tots.nephewSelf += block.dat.rewardNephew
			tots.minedSelf += float64(block.dat.rewardMined)
		} else {
			tots.nephewHonest += block.dat.rewardNephew
			tots.minedHonest += float64(block.dat.rewardMined)
		}
	}
	tots.totalSelf = tots.minedSelf + tots.nephewSelf + tots.uncleSelf
	tots.totlaHonest = tots.minedHonest + tots.nephewHonest + tots.uncleHonest
	tots.total = tots.totalSelf + tots.totlaHonest

	tots.absSelfRev = tots.totalSelf / (tots.minedSelf + tots.minedHonest)
	tots.absHonestRev = tots.totlaHonest / (tots.minedSelf + tots.minedHonest)

	return &tots
}

func uncleReward(block *block, tots *chainTots) {
	for _, b := range block.uncleBlocks {
		if b.dat.selfish {
			tots.uncleSelf += b.dat.rewardUncle
		} else {
			tots.uncleHonest += b.dat.rewardUncle
		}
	}
}

func (c *chainTots) String() string {
	return fmt.Sprintf(
		"Total: %f\t TotalSelf: %f, TotalHonest: %f\nuncleSelf: %f\tUncleHonest: %f\nNephewSelf: %f\tNephewHonest: %f\nAbsSelf: %f\tAbsHonest: %f",
		c.total, c.totalSelf, c.totlaHonest,
		c.uncleSelf, c.uncleHonest,
		c.nephewSelf, c.nephewHonest,
		c.absSelfRev, c.absHonestRev,
	)
}
