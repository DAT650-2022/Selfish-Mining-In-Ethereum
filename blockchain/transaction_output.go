package main

import (
	"bytes"
	"fmt"
)

// TXOutput represents a transaction output
type TXOutput struct {
	Value      int    // The transaction value
	PubKeyHash []byte // The conditions to claim this output. For this demo we will use the hash of the public key (used to "lock" the output)
}

// Lock locks the transaction to a specific address
// Only this address owns this transaction
func (out *TXOutput) Lock(address string) {
	base58 := Base58Decode([]byte(address))
	out.PubKeyHash = base58[1 : len(base58)-addressChecksumLen]
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(pubKeyHash, out.PubKeyHash)
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	txOut := TXOutput{Value: value}
	txOut.Lock(address)
	return &txOut
}

func (out TXOutput) String() string {
	return fmt.Sprintf("{%d, %x}", out.Value, out.PubKeyHash)
}
