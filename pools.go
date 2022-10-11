package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"
)

type poolRewards struct {
	totalReward int
	rewards     []*dataUnit
}

type system struct {
	bc          *blockchain
	privChain   []*block
	fork        bool
	forkSelfish bool
	forkDepth   int
	fo          map[int][]*block
}

func newSystem() *system {
	return &system{}
}

var sys = newSystem()
var privChain []*block

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

	go honestPool(55, honestBlockChan, honestnetCom)
	go selfishPool(45, selfBlockChan, selfishnetCom)

	// Network power of the selfish pool
	for {
		select {
		case b := <-selfBlockChan:
			sys.addBlock(b, true)
			selfishBlocks += 1
			fmt.Printf("Selfish published depth: %d\n", b.depth)
			fmt.Println("___________________________")
			fmt.Println(sys.bc.String())
		case b := <-honestBlockChan:
			sys.addBlock(b, false)
			honestBlocks += 1
			fmt.Printf("Honest publish depth: %d\n", b.depth)
			fmt.Println("___________________________")
			fmt.Println(sys.bc.String())
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func selfishPool(power int, blockCom chan *block, netCom chan int) {
	privChain = []*block{}
	potUncles := []int{} // indexes of potiential uncle blocks from publick chain.
	for {
		// The selfish pool mines a new block
		time.Sleep(200 * time.Millisecond)

		randroll := rand.Intn(100)
		if power >= randroll {
			// We succeded, new block
			nb := createBlock(power, true)
			// If we already have a private chain -> Parent and depth follow private chain instead of public blockchain
			if len(privChain) > 0 {
				nb.parent = privChain[len(privChain)-1]
				nb.parentHash = privChain[len(privChain)-1].hash
				nb.depth = privChain[len(privChain)-1].depth + 1

			}
			// Check if we have any potentiall uncleblocks
			if len(sys.bc.uncles) > 0 {
				uncs := findUncles(&potUncles, nb.depth)
				// uncs is list of index(in chain) to valid uncles
				// returns max 2
				if len(uncs) > 0 {
					for _, u := range uncs {
						nb.uncleBlocks = append(nb.uncleBlocks, sys.bc.chain[u])

					}
				}
			}
			nb.calckRewards()
			privChain = append(privChain, nb)
			fmt.Println(fmt.Sprintf("Selfish: new secret block added depth: %d", nb.depth))
		}

		// Some honest miners has mined a block and we have private blocks
		if len(privChain) > 0 && sys.bc.CurrentBlock().depth >= privChain[0].depth {
			// 1. The miner references all (unreferenced) uncle blocks based on its public branches

			// Honest pool is ahead of us. Scrap private chain and mine on new block.
			if privChain[len(privChain)-1].depth < sys.bc.CurrentBlock().depth {
				privChain = []*block{}
				fmt.Println("Selfish: abandon")
				// If we are tied with honest pool. Release the last block in the private branch and scrap.
			} else if privChain[len(privChain)-1].depth == sys.bc.CurrentBlock().depth {
				blockCom <- privChain[len(privChain)-1]
				privChain = []*block{}
				fmt.Println("Selfish: Tied release")
				// If we are ahead by only one. Publish private branch.
			} else if privChain[len(privChain)-1].depth == sys.bc.CurrentBlock().depth+1 {
				for _, block := range privChain {
					blockCom <- block
				}
				privChain = []*block{}
				fmt.Println("Full release")
				// If we are ahead by more than 2. Release block until we reach public chain
			} else if privChain[len(privChain)-1].depth >= sys.bc.CurrentBlock().depth+2 {
				//toRelease := privChain[len(privChain)-1].depth - (sys.bc.CurrentBlock().depth + 2)
				toRelease := sys.bc.CurrentBlock().depth - privChain[0].depth + 1 // same depth = 0, release one?, one ahead = 1 relase 2?
				for i := 0; i < toRelease; i++ {
					blockCom <- privChain[i]
				}

				privChain = privChain[toRelease:]
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
	depth := sys.bc.CurrentBlock().depth + 1
	if selfish {
		dat = dataUnit{selfish: true}
	} else {
		dat = dataUnit{selfish: false}
		if sys.fork && !sys.forkSelfish {
			depth = sys.fo[sys.forkDepth][len(sys.fo[sys.forkDepth])-1].depth + 1
		}
	}
	newBlock := block{
		hash:       []byte(randomString(10)),
		parent:     sys.bc.CurrentBlock(),
		dat:        dat,
		parentHash: sys.bc.CurrentBlock().hash,
		depth:      depth,
	}
	return &newBlock
}

func (s *system) addBlock(b *block, selfish bool) {
	if !s.fork {
		if b.depth == s.bc.CurrentBlock().depth && !s.fork { // New fork has appeard
			println("NEW FORK!!!!")
			s.fork = true
			s.forkDepth = b.depth
			s.forkSelfish = b.dat.selfish
			s.fo = make(map[int][]*block)
			s.fo[b.depth] = []*block{b}
			return
		} else if b.depth < s.bc.CurrentBlock().depth {
			// find parent
			println("SOmethgin wong")
			k := b
			i := 0
			for {
				if bytes.Equal(k.parentHash, sys.bc.chain[k.parent.depth].hash) { // Parent is in main chain
					if _, ok := sys.bc.uncles[k.depth]; !ok {
						sys.bc.uncles[k.depth] = k
						break
					}
				}
				k = k.parent
				i++
				if i == 3 {
					break
				}
			}
			return
		}

		s.bc.addNewBlock(b)
	}

	if !b.dat.selfish && s.fork && b.depth == s.bc.CurrentBlock().depth+1 && len(privChain) <= 0 { // selfish realesed tie
		if s.forkSelfish {
			s.bc.addNewBlock(b)
			sys.bc.uncles[s.forkDepth] = s.fo[s.forkDepth][0]
		} else {
			sys.bc.uncles[s.forkDepth] = sys.bc.chain[s.forkDepth]
			s.fo[s.forkDepth] = append(s.fo[s.forkDepth], b)
			s.bc.chain = s.bc.chain[:s.forkDepth]
			for _, bl := range s.fo[s.forkDepth] {
				s.bc.addNewBlock(bl)
			}
			delete(s.fo, b.depth)
		}
		s.fork = false
		println("Fork solved: case 1")
		return
	} else if !b.dat.selfish && s.fork && b.depth == s.bc.CurrentBlock().depth+1 && len(privChain) > 0 { // selfish realesed tie
		if s.forkSelfish {
			s.bc.addNewBlock(b)
		} else {
			s.fo[s.forkDepth] = append(s.fo[s.forkDepth], b)
		}
		println("Fork: case 1.2")
		return
	} else if b.dat.selfish && s.fork && b.depth == s.bc.CurrentBlock().depth+1 { // selfish realesed tie
		if !s.forkSelfish {
			s.bc.addNewBlock(b)
			sys.bc.uncles[s.forkDepth] = s.fo[s.forkDepth][0]
		} else {
			sys.bc.uncles[s.forkDepth] = sys.bc.chain[s.forkDepth]
			s.fo[s.forkDepth] = append(s.fo[s.forkDepth], b)
			s.bc.chain = s.bc.chain[:s.forkDepth]
			for _, bl := range s.fo[s.forkDepth] {
				s.bc.addNewBlock(bl)
			}
			// s.bc.chain = append(s.bc.chain, s.fo[s.forkDepth]...)
			delete(s.fo, b.depth)
		}
		s.fork = false
		println("Fork solved: case 2")
		return
	}

	if !b.dat.selfish && b.depth == s.bc.CurrentBlock().depth && s.fork {
		if !s.forkSelfish {
			s.fo[s.forkDepth] = append(s.fo[s.forkDepth], b)
		}
	} else if b.dat.selfish && b.depth == s.bc.CurrentBlock().depth && s.fork {
		if s.forkSelfish {
			s.fo[s.forkDepth] = append(s.fo[s.forkDepth], b)
		}
	}

	if b.depth < s.bc.CurrentBlock().depth && s.fork {
		if s.forkSelfish {
			if b.dat.selfish {
				s.fo[s.forkDepth] = append(s.fo[s.forkDepth], b)
			}
		} else {
			if !b.dat.selfish {
				s.fo[s.forkDepth] = append(s.fo[s.forkDepth], b)
			}
		}
	}

}

func honestPool(power int, blockCom chan *block, netCom chan int) {
	// missedBlocks := 0
	for {
		time.Sleep(200 * time.Millisecond)
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
