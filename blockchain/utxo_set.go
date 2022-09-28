package main

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// UTXOSet represents a set of UTXO as an in-memory cache
// The key of the most external map is the transaction ID
// (encoded as string) that contains these outputs
// {map of transaction ID -> {map of TXOutput Index -> TXOutput}}
type UTXOSet map[string]map[int]TXOutput

// FindSpendableOutputs finds and returns unspent outputs in the UTXO Set
// to reference in inputs
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulatedAmount := 0
	for ID, outputs := range u {
		var unspent []int
		for i, out := range outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				if accumulatedAmount < amount {
					accumulatedAmount += out.Value
					unspent = append(unspent, i)
				}
			}
		}
		if unspent != nil {
			unspentOutputs[ID] = unspent
		}
	}
	return accumulatedAmount, unspentOutputs
}

// FindUTXO finds all UTXO in the UTXO Set for a given unlockingData key (e.g., address)
// This function ignores the index of each output and returns
// a list of all outputs in the UTXO Set that can be unlocked by the user
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXO []TXOutput
	for _, out := range u {
		for _, i := range out {
			if i.IsLockedWithKey(pubKeyHash) {
				UTXO = append(UTXO, i)
			}
		}
	}
	return UTXO
}

// CountUTXOs returns the number of transactions outputs in the UTXO set
func (u UTXOSet) CountUTXOs() int {
	count := 0
	for _, out := range u {
		count += len(out)
	}
	return count
}

// Update updates the UTXO Set with the new set of transactions
func (u UTXOSet) Update(transactions []*Transaction) {
	for _, tx := range transactions {
		if tx.IsCoinbase() == false {
			for _, Vin := range tx.Vin {
				if ele, ok := u[hex.EncodeToString(Vin.Txid)]; ok {
					outPutMap := make(map[int]TXOutput)
					for i, out := range ele {
						if i != Vin.OutIdx {
							outPutMap[i] = out
						}
					}
					if len(outPutMap) == 0 {
						delete(u, hex.EncodeToString(Vin.Txid))
					} else {
						u[hex.EncodeToString(Vin.Txid)] = outPutMap
					}
				}
			}
		}

		newOutPuts := make(map[int]TXOutput)
		for _, output := range tx.Vout {
			newOutPuts[len(newOutPuts)] = output
		}
		u[hex.EncodeToString(tx.ID)] = newOutPuts
	}
}

func (u UTXOSet) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- UTXO SET:"))
	for txid, outputs := range u {
		lines = append(lines, fmt.Sprintf("     TxID: %s", txid))
		for i, out := range outputs {
			lines = append(lines, fmt.Sprintf("           Output %d: %v", i, out))
		}
	}

	return strings.Join(lines, "\n")
}
