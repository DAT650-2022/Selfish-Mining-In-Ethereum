package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	const RUNS = 5
	const BLOCKS = 1000

	start := time.Now()
	for r := 1; r < RUNS; r++ {
		chains := []*blockchain{}
		comChans := make([]chan *blockchain, 0)
		for i := 1; i < 10; i += 1 {
			selfPow := (i) * 5 * 10
			for j := 0; j < 2; j++ {
				dyn := j == 0
				com := make(chan *blockchain)
				honestPow := 1000 - selfPow
				go poolController(honestPow, selfPow, BLOCKS, dyn, com, fmt.Sprintf("chain-%d-%d_selfPow-%d_dyn-%t", i, r, selfPow, dyn))
				comChans = append(comChans, com)
				println("Started new: ", selfPow)
			}
		}
		for i := 0; i < len(comChans); i++ {
			chains = append(chains, <-comChans[i])
		}
		for _, C := range chains {
			chainRewardToCsv(C, C.name)
		}
	}

	runtime := time.Since(start)
	log.Printf("Total runtime took %s", runtime)

	println("DONE!")
	var inp string
	fmt.Scanln(&inp)
}
