package main

func main() {
	rewards := make(chan poolRewards)
	chain := &blockchain{} // Probably going to be moved to pools.go
	go poolController(1000, chain, rewards)

	// main loop
	for {
		// print data and potentially accept user input
		// like controlling a tickrate
		printData()
		break
	}
	saveTofile()

}

func printData() {
	// print some data to consoel.
}

func saveTofile() {
	// Save data to csv file.
	// function might be moved to another location.
}
