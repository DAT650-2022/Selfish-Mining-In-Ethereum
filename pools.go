package main

import (
	"math/rand"
	"time"
)

type poolRewards struct {
	totalReward int
	rewards     []*dataUnit
}

func poolController(com chan poolRewards) {
	chain := newBlockChain()

	selfBlockChan := make(chan *block)
	honestBlockChan := make(chan *block)
	// Used to simulate network to be able to inform
	// "network" of a new block being published.
	netCom := make(chan int, 100)

	go selfishPool(45, selfBlockChan, netCom)
	go honestPool(55, honestBlockChan, netCom)

	// main loop
	for {
		select {
		case b := <-selfBlockChan:
			chain.addNewBlock(b)
			// netCom <- 1
		case b := <-honestBlockChan:
			chain.addNewBlock(b)
			println("\tHonest: created new block")
			netCom <- 1
		default:
			time.Sleep(50 * time.Millisecond)
		}

	}
}

// TODO: Add a 'local' version of the blockchain.
// May be better, maybe not. Would be closer to reality.
func selfishPool(power int, blockCom chan *block, netCom chan int) {
	privChain := make([]*block, 0)
	pubchain := 0
	for {
		// do work
		time.Sleep(1 * time.Second)
		if rand.Intn(16) == 15 {
			// We succeded, new block
			nb := selfishMiner(power)
			privChain = append(privChain, nb)
			println("Selfish: new secret block added")
		}

		if len(netCom) == 1 && len(privChain) > 0 {
			// they have released a block, time to reveal our own
			blockCom <- privChain[0]
			println("Selfish: Published secret block")
			privChain = privChain[1:] // removes the first element.
			<-netCom
			continue
		} else if len(netCom) == 0 {
			continue
		} else { // there is more than one block added since last check.
			for _ = range netCom {
				pubchain += 1
			}
		}
		if pubchain == len(privChain)-1 {
			// they have caught up, release all if any private held blocks
			for _, b := range privChain {
				blockCom <- b
			}
			pubchain = 0
			privChain = make([]*block, 0)
			println("Selfish: full release")
		} else if pubchain >= len(privChain)-1 {
			// they have passed or have more blocks, join the publick chain
			pubchain = 0
			privChain = make([]*block, 0)
			println("Selfish: abandon")
		} else if pubchain <= len(privChain)-2 {
			// they are close but not fully caught up.
			// Need to release blocks.
			// TODO: this is not technically part of the algo in paper, probably not needed.
			println("Selfish: TODO")
			continue
		}

	}
}
func selfishMiner(power int) *block {
	// Creates a dataunit of expected value.
	// actual final values gets calculated from the final blockchain.
	newBlock := block{
		dat: dataUnit{selfish: true},
	}
	return &newBlock
}

func honestPool(power int, blockCom chan *block, netCom chan int) {
	// missedBlocks := 0
	for {
		time.Sleep(1 * time.Second)
		if rand.Intn(16) != 15 {
			continue
		}

		// We have mined a block, time to publish
		// check if our block is relevant.
		// if len(netCom) <= 1 {
		// 	// Our block no good
		// 	// might be a uncle block
		// 	// TODO: still send our block to blockcom
		// 	// the poolcontroller should handle rest?
		// 	// for now empty queue and just continue
		// 	for _ = range netCom {
		// 		missedBlocks += 1
		// 	}
		// 	continue
		// }
		// We have a new blokc, publish.
		blockCom <- honsetMiner()
	}
}
func honsetMiner() *block {
	// Creates a dataunit of expected value.
	// actual final values gets calculated from the final blockchain.
	// Dataunit is probably going to be set by poolController.
	// TODO: datunit.
	newBlock := block{
		dat: dataUnit{selfish: false},
	}
	return &newBlock
}

// For early early tests we just run a 1/15 chance
// for succes, the miners run this work 1 time second.
// +- some value to adjust for hashing power.
// TODO: Fully create it later.
func doWork() bool {
	return rand.Intn(16) == 15
}
