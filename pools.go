package main

import (
	"fmt"
	"math/rand"
	"time"
)

type poolRewards struct {
	totalReward int
	rewards     []*dataUnit
}

type system struct {
	bc *blockchain
}

func newSystem() *system {
	return &system{}
}

func poolController(com chan poolRewards) {
	system := newSystem()
	system.bc = newBlockChain()

	selfBlockChan := make(chan *block)
	honestBlockChan := make(chan *block)
	// Used to simulate network to be able to inform
	// "network" of a new block being published.
	selfishnetCom := make(chan int, 100)
	honestnetCom := make(chan int, 100)

	go honestPool(55, honestBlockChan, honestnetCom)
	go selfishPool(45, selfBlockChan, selfishnetCom)

	// Network power of the selfish pool
	//gamma := 0.5
	// main loop
	for {
		select {
		case b := <-selfBlockChan:
			system.addBlock(b, true)
			honestnetCom <- len(system.bc.chain)
			fmt.Println("___________________________")
			fmt.Println(system.bc.String())
		case b := <-honestBlockChan:
			system.addBlock(b, false)
			selfishnetCom <- len(system.bc.chain)
			fmt.Println("___________________________")
			fmt.Println(system.bc.String())
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// TODO: Add a 'local' version of the blockchain.
// May be better, maybe not. Would be closer to reality.
func selfishPool(power int, blockCom chan *block, netCom chan int) {
	var privChain []*block
	for {
		// The selfish pool mines a new block
		time.Sleep(1 * time.Second)
		randroll := rand.Intn(100)
		if power >= randroll {
			// We succeded, new block
			nb := s.createBlock(power, true)
			// If we already have a private chain -> Parent and depth follow private chain instead of public blockchain
			if len(privChain) > 0 {
				nb.parent = privChain[len(privChain)-1]
				nb.parentHash = privChain[len(privChain)-1].hash
				nb.depth = privChain[len(privChain)-1].depth + 1
			}
			privChain = append(privChain, nb)
			fmt.Println("Selfish: new secret block added")
			// When we are ahead by one and our private chain has 2 blocks -> Limited advantage
			if len(privChain) == 2 && privChain[len(privChain)-1].depth-1 == s.bc.CurrentBlock().depth {
				fmt.Println("Selfish: Ahead by 1 -> Full release ")
				for _, block := range privChain {
					blockCom <- block
				}
				privChain = []*block{}
			}
		}
		// Some honest miners has mined a block and we have private blocks
		if len(netCom) > 0 && len(privChain) > 0 {
			// 1. The miner references all (unreferenced) uncle blocks based on its public branches

			// Honest pool is ahead of us. Scrap private chain and mine on new block.
			if privChain[len(privChain)-1].depth < s.bc.CurrentBlock().depth {
				privChain = []*block{}
				fmt.Println("Selfish: abandon")
				// If we are tied with honest pool. Release the last block in the private branch and scrap.
			} else if privChain[len(privChain)-1].depth == s.bc.CurrentBlock().depth {
				blockCom <- privChain[len(privChain)-1]
				privChain = []*block{}
				fmt.Println("Selfish: Tied release")
				// If we are ahead by one. Publish private branch.
			} else if privChain[len(privChain)-1].depth == s.bc.CurrentBlock().depth+1 {
				for _, block := range privChain {
					blockCom <- block
				}
				privChain = []*block{}
				fmt.Println("Full release")
				// If we are ahead by more than 2. Release block until we reach public chain
			} else if privChain[len(privChain)-1].depth >= s.bc.CurrentBlock().depth+2 {
				toRelease := privChain[len(privChain)-1].depth - (s.bc.CurrentBlock().depth + 2)
				for _, block := range privChain[:toRelease] {
					blockCom <- block
				}
				privChain = privChain[toRelease:]
				fmt.Println("Selfish: Ahead by 2 or more")
			}
			<-netCom
		}
	}
}

func createBlock(power int, selfish bool) *block {
	// Creates a dataunit of expected value.
	// actual final values gets calculated from the final blockchain.
	var dat dataUnit
	if selfish {
		dat = dataUnit{selfish: true}
	} else {
		dat = dataUnit{selfish: false}
	}
	newBlock := block{
		hash:       []byte(randomString(10)),
		dat:        dat,
		parentHash: s.bc.CurrentBlock().hash,
		depth:      s.bc.CurrentBlock().depth + 1,
	}
	return &newBlock
}

func (s *system) addBlock(block *block, selfish bool) {
	gamma := 50
	// If two blocks are on the same depth -> Fork
	if s.bc.CurrentBlock().depth == block.depth {
		randroll := rand.Intn(100)
		// If selfish pool reaches most nodes because of network power
		if randroll >= gamma {
			if selfish == true {
				s.bc.chain = s.bc.chain[:len(s.bc.chain)-1]
				s.bc.addNewBlock(block)
				return
			} else {
				return
			}
		}
	}
	s.bc.addNewBlock(block)
}

func honestPool(power int, blockCom chan *block, netCom chan int) {
	// missedBlocks := 0
	for {
		time.Sleep(1 * time.Second)
		if power >= rand.Intn(100) {
			blockCom <- createBlock(power, false)
			continue
		}

	}
}

// For early early tests we just run a 1/15 chance
// for succes, the miners run this work 1 time second.
// +- some value to adjust for hashing power.
// TODO: Fully create it later.
func doWork() bool {
	return rand.Intn(16) == 15
}

func randomString(length int) string {
	b := make([]byte, length)
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	charSet := "aAbBcCdDeEfFgGhHiIjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyYzZ"
	for i := range b {
		b[i] = charSet[seededRand.Intn(len(charSet)-1)]
	}
	return string(b)
}
