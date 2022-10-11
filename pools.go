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
	bc        *blockchain
	privChain []*block
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
	netCom := make(chan int, 100)

	go system.selfishPool(70, selfBlockChan, netCom)
	go system.honestPool(30, honestBlockChan, netCom)

	// Network power of the selfish pool
	//gamma := 0.5
	// main loop
	for {
		select {
		case b := <-selfBlockChan:
			system.addBlock(b, true)
			// netCom <- 1
			fmt.Println("___________________________")
			fmt.Println(system.bc.String())
		case b := <-honestBlockChan:
			system.addBlock(b, false)
			fmt.Println("Honest: created new block")
			netCom <- 1
			fmt.Println("___________________________")
			fmt.Println(system.bc.String())
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// TODO: Add a 'local' version of the blockchain.
// May be better, maybe not. Would be closer to reality.
func (s *system) selfishPool(power int, blockCom chan *block, netCom chan int) {
	s.privChain = []*block{}
	for {
		// The selfish pool mines a new block
		time.Sleep(1 * time.Second)
		randroll := rand.Intn(100)
		if power >= randroll {
			// We succeded, new block
			nb := s.createBlock(power, true)
			// If we already have a private chain -> Parent and depth follow private chain instead of public blockchain
			if len(s.privChain) > 0 {
				nb.parent = s.privChain[len(s.privChain)-1]
				nb.parentHash = s.privChain[len(s.privChain)-1].hash
				nb.depth = s.privChain[len(s.privChain)-1].depth + 1
			}
			s.privChain = append(s.privChain, nb)
			fmt.Println("Selfish: new secret block added")
			// When we are ahead by one and our private chain has 2 blocks -> Limited advantage
			if len(s.privChain) == 2 && s.privChain[len(s.privChain)-1].depth-1 == s.bc.CurrentBlock().depth {
				fmt.Println("Selfish: Ahead by 1 -> Limited advantage release")
				for _, block := range s.privChain {
					blockCom <- block
				}
				s.privChain = []*block{}
			}
		}
		// Some honest miners has mined a block and we have private blocks
		if len(netCom) > 0 && len(s.privChain) > 0 {
			// 1. The miner references all (unreferenced) uncle blocks based on its public branches

			// Honest pool is ahead of us. Scrap private chain and mine on new block.
			if s.privChain[len(s.privChain)-1].depth < s.bc.CurrentBlock().depth {
				s.privChain = []*block{}
				fmt.Println("Selfish: Behind main chain -> Abandon")
				// If we are tied with honest pool. Release the last block in the private branch and scrap.
			} else if s.privChain[len(s.privChain)-1].depth == s.bc.CurrentBlock().depth {
				blockCom <- s.privChain[len(s.privChain)-1]
				s.privChain = []*block{}
				fmt.Println("Selfish: Same depth as main chain -> Tied release")
				// If we are ahead by one. Publish private branch.
			} else if s.privChain[len(s.privChain)-1].depth == s.bc.CurrentBlock().depth+1 {
				for _, block := range s.privChain {
					blockCom <- block
				}
				s.privChain = []*block{}
				fmt.Println("Selfish: Ahead by one -> Full release")
				// If we are ahead by more than 2. Release block until we reach public chain
			} else if s.privChain[len(s.privChain)-1].depth >= s.bc.CurrentBlock().depth+2 {
				toRelease := s.bc.CurrentBlock().depth - s.privChain[0].depth
				for _, block := range s.privChain[:toRelease] {
					blockCom <- block
				}
				s.privChain = s.privChain[toRelease:]
				fmt.Println("Selfish: Ahead by 2 or more -> Catch up")
			}
			<-netCom
		}
	}
}

func (s *system) createBlock(power int, selfish bool) *block {
	// Creates a dataunit of expected value.
	// actual final values gets calculated from the final blockchain.
	var dat dataUnit
	if selfish {
		dat = dataUnit{selfish: true}
	} else {
		dat = dataUnit{selfish: false}
	}
	newBlock := block{
		hash:        []byte(randomString(10)),
		parent:      s.bc.CurrentBlock(),
		parentHash:  s.bc.CurrentBlock().hash,
		uncleBlocks: []*block{},
		dat:         dat,
		depth:       s.bc.CurrentBlock().depth + 1,
	}
	return &newBlock
}

func (s *system) addBlock(block *block, selfish bool) {
	// If current selfish block is at same depth as public chain. Possible fork has occurred ->
	// Check if selfish chain is longer than public chain. Adopt the longest chain. If not -> Roll based on network power
	if selfish {
		if block.depth <= s.bc.CurrentBlock().depth {
			if len(s.privChain) >= (s.bc.CurrentBlock().depth - block.depth) {
				s.bc.chain = s.bc.chain[:len(s.bc.chain)-1]
				s.bc.addNewBlock(block)
				fmt.Println(block)
			} else {
				gamma := 50
				randroll := rand.Intn(100)
				if randroll >= gamma {
					s.bc.chain = s.bc.chain[:len(s.bc.chain)-1]
					s.bc.addNewBlock(block)
				}
			}
		} else {
			s.bc.addNewBlock(block)
		}
	}

	if !selfish {
		s.bc.addNewBlock(block)
	}
}

func (s *system) honestPool(power int, blockCom chan *block, netCom chan int) {
	// missedBlocks := 0
	for {
		time.Sleep(1 * time.Second)
		if power >= rand.Intn(100) {
			blockCom <- s.createBlock(power, false)
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
