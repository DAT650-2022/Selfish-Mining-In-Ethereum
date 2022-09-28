package main

type block struct {
	hash          []byte
	prevBlock     *block
	prevBlockHash []byte
	uncleBlocks   []*block
	difficulty    int
}
