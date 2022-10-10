package main

import (
	"time"
)

func main() {
	rewards := make(chan poolRewards)

	go poolController(rewards)

	// main loop
	for {
		// print data and potentially accept user input
		// like controlling a tickrate
		printData()
		break
	}
	saveTofile()

	time.Sleep(120 * time.Second)
}

func printData() {
	// print some data to consoel.
}

func saveTofile() {
	// Save data to csv file.
	// function might be moved to another location.
}
