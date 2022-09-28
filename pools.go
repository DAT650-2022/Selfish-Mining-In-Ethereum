package main

type dataUnit struct {
	rewardTot    int
	rewardMined  int
	rewardNephew int
	rewardUncle  int
	selfish      bool
}

type poolRewards struct {
	totalReward int
	rewards     []*dataUnit
}

func poolController(nodes int, chain *blockchain, com chan poolRewards) {
	selfishChan := make(chan dataUnit)
	honestChan := make(chan dataUnit)

	go selfishPool(nodes, selfishChan)
	go honestPool(nodes, honestChan)

}

func selfishPool(nodes int, com chan dataUnit) {
	for {
		// run the algorithm
		return
	}
}

func honestPool(nodes int, com chan dataUnit) {
	for {
		// run the algorithm
		return
	}
}
