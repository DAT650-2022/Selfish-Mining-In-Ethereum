package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

func ChainTotxt(bc *blockchain, name string) {
	path, _ := filepath.Abs(fmt.Sprintf("./%s.txt", name))
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	w := bufio.NewWriter(file)

	for i := 0; i < len(bc.chain); i++ {
		b := bc.chain[i]
		w.WriteString(
			fmt.Sprintf(
				"|Bhash:\t%s\tPhash:\t%s\tSelfish:%t\t|Uncles: %d\n",
				b.hash, b.parentHash, b.dat.selfish, len(b.uncleBlocks),
			))

	}

	w.Flush()
}

func chainRewardToCsv(bc *blockchain, name string) {
	path, _ := filepath.Abs(fmt.Sprintf("./data/%s.csv", name))
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	w := bufio.NewWriter(file)

	rew := calcChainRewards(bc)

	w.WriteString("Chain-number-run;Blocks;Total;TotalSelfish;TotalHonest;UncleSelfish;UncleHonest;NephewSelfish;NephewHonest;MinedSelf;MinedHonest;absSelfRev;absHonestRev\n")
	w.WriteString(
		fmt.Sprintf(
			"%s;%d;%f;%f;%f;%f;%f;%f;%f;%f;%f;%f;%f\n",
			bc.name, len(bc.chain),
			rew.total, rew.totalSelf, rew.totlaHonest,
			rew.uncleSelf, rew.uncleHonest,
			rew.nephewSelf, rew.nephewHonest,
			rew.minedSelf, rew.minedHonest,
			rew.absSelfRev, rew.absHonestRev,
		))

	w.Flush()
}
