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

var sys = newSystem()

func poolController(com chan poolRewards) {
	sys.bc = newBlockChain()

	selfishBlocks := 0
	honestBlocks := 0

	selfBlockChan := make(chan *block)
	honestBlockChan := make(chan *block)
	// Used to simulate network to be able to inform
	// "network" of a new block being published.
	selfishnetCom := make(chan int, 100)
	honestnetCom := make(chan int, 100)

	go honestPool(20, honestBlockChan, honestnetCom)
	go sys.selfishPool(80, selfBlockChan, selfishnetCom)

	// Network power of the selfish pool
	//gamma := 0.5
	// main loop
	for {
		select {
		case b := <-selfBlockChan:
			sys.addBlock(b, true)
			honestnetCom <- len(sys.bc.chain) - 1
			selfishBlocks += 1
			fmt.Println("___________________________")
			fmt.Println(sys.bc.String())
			//fmt.Println("Selfish published")
		case b := <-honestBlockChan:
			sys.addBlock(b, false)
			selfishnetCom <- len(sys.bc.chain) - 1
			honestBlocks += 1
			fmt.Println("___________________________")
			fmt.Println(sys.bc.String())
			fmt.Println("Honest publish")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (s *system) selfishPool(power int, blockCom chan *block, netCom chan int) {
	s.privChain = []*block{}
	potUncles := []int{} // indexes of potiential uncle blocks from publick chain.
	for {
		// The selfish pool mines a new block
		time.Sleep(1 * time.Second)

		randroll := rand.Intn(100)
		if power >= randroll {
			// We succeded, new block
			nb := createBlock(power, true)
			// If we already have a private chain -> Parent and depth follow private chain instead of public blockchain
			if len(s.privChain) > 0 {
				nb.parent = s.privChain[len(s.privChain)-1]
				nb.parentHash = s.privChain[len(s.privChain)-1].hash
				nb.depth = s.privChain[len(s.privChain)-1].depth + 1

			}
			// Check if we have any potentiall uncleblocks
			if len(potUncles) > 0 {
				uncs := findUncles(&potUncles, nb.depth)
				// uncs is list of index(in chain) to valid uncles
				// returns max 2
				if len(uncs) > 0 {
					for u := range uncs {
						nb.uncleBlocks = append(nb.uncleBlocks, sys.bc.chain[u])
					}
				}
			}
			nb.calckRewards()
			s.privChain = append(s.privChain, nb)
			fmt.Println("Selfish: new secret block added")
		}

		// Some honest miners has mined a block and we have private blocks
		if len(s.privChain) > 0 && sys.bc.CurrentBlock().depth >= s.privChain[0].depth {
			// 1. The miner references all (unreferenced) uncle blocks based on its public branches

			// Honest pool is ahead of us. Scrap private chain and mine on new block.
			if s.privChain[len(s.privChain)-1].depth < sys.bc.CurrentBlock().depth {
				s.privChain = []*block{}
				fmt.Println("Selfish: abandon")
				// If we are tied with honest pool. Release the last block in the private branch and scrap.
			} else if s.privChain[len(s.privChain)-1].depth == sys.bc.CurrentBlock().depth {
				blockCom <- s.privChain[len(s.privChain)-1]
				s.privChain = []*block{}
				fmt.Println("Selfish: Tied release")
				// If we are ahead by only one. Publish private branch.
			} else if s.privChain[len(s.privChain)-1].depth == sys.bc.CurrentBlock().depth+1 {
				for _, block := range s.privChain {
					blockCom <- block
				}
				s.privChain = []*block{}
				fmt.Println("Full release")
				// If we are ahead by more than 2. Release block until we reach public chain
			} else if s.privChain[len(s.privChain)-1].depth >= sys.bc.CurrentBlock().depth+2 {
				//toRelease := privChain[len(privChain)-1].depth - (sys.bc.CurrentBlock().depth + 2)
				toRelease := sys.bc.CurrentBlock().depth - s.privChain[0].depth + 1 // same depth = 0, release one?, one ahead = 1 relase 2?
				for i := 0; i < toRelease; i++ {
					blockCom <- s.privChain[i]
				}

				s.privChain = s.privChain[toRelease:]
				fmt.Println("Selfish: Ahead by 2 or more")
			}
		}
	}
}

// Returns index of one or more uncleblocks
func findUncles(uncs *[]int, depth int) []int {
	response := []int{}
	for i := 0; i < len(*uncs); i++ {
		if depth-(*uncs)[i] <= 6 { // Can't be over 6 blocks old
			// TODO: Can't remeber if its supposed to ignore, or its just no reward for older than 6
			response = append(response, (*uncs)[i])
			// Remove used index from slice
			*uncs = append((*uncs)[:i], (*uncs)[i+1:]...)
		}
		if len(response) >= 2 {
			return response
		}

	}

	if len(response) == 0 && len(*uncs) > 0 {
		// Since none was used, they are all to old, jsut create new list
		*uncs = []int{}
	}
	return response
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
		parent:     sys.bc.CurrentBlock(),
		dat:        dat,
		parentHash: sys.bc.CurrentBlock().hash,
		depth:      sys.bc.CurrentBlock().depth + 1,
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

func honestPool(power int, blockCom chan *block, netCom chan int) {
	// missedBlocks := 0
	for {
		time.Sleep(1 * time.Second)
		if power >= rand.Intn(100) {
			nb := createBlock(power, false)
			nb.calckRewards()
			blockCom <- nb

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
