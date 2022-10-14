package main

import (
	"os"
	"strconv"
)

func main() {
	selfishPower, _ := strconv.Atoi(os.Args[1])
	rewards := make(chan *blockchain)

	go poolController(rewards, selfishPower)

	// main loop
	for {
		// print data and potentially accept user input
		// like controlling a tickrate
		printData()
		break
	}
	saveTofile()

	// time.Sleep(20000 * time.Second)
	chain := <-rewards

	chainRewardToCsv(chain, os.Args[1])
	totals := calcChainRewards(chain)

	print(totals)

	println("DONE!")
}

func printData() {
	// print some data to consoel.
	// for _, b := range sys.bc.chain {
	// 	println("==========")
	// 	print(b)
	// 	println("==========")
	// }
}

func saveTofile() {
	// Save data to csv file.
	// function might be moved to another location.
}
