package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

var (
	ErrNoFunds         = errors.New("not enough funds")
	ErrTxInputNotFound = errors.New("transaction input not found")
)

// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to, data string) (*Transaction, error) {
	if data == "" {
		token := make([]byte, 20)
		_, err := rand.Read(token)
		if err != nil {
			return nil, err
		}
		data = fmt.Sprintf(string(token))
	}
	tx := Transaction{ID: nil, Vin: []TXInput{{Txid: nil, OutIdx: -1, Signature: nil, PubKey: []byte(data)}}, Vout: []TXOutput{*NewTXOutput(BlockReward, to)}}
	tx.ID = tx.Hash()
	return &tx, nil
}

// NewUTXOTransaction creates a new UTXO transaction
// NOTE: The returned tx is NOT signed!
func NewUTXOTransaction(pubKey []byte, to string, amount int, utxos UTXOSet) (*Transaction, error) {
	pubKeyHash := GetPubKeyHashFromAddress(string(GetAddress(pubKey)))
	accumulatedAmount, output := utxos.FindSpendableOutputs(pubKeyHash, amount)
	if accumulatedAmount < amount {
		return nil, ErrNoFunds
	}

	var inputs []TXInput
	for txID, out := range output {
		for _, idx := range out {
			inputs = append(inputs, TXInput{Txid: Hex2Bytes(txID), OutIdx: idx, Signature: nil, PubKey: pubKey})
		}
	}

	var outputs []TXOutput
	change := accumulatedAmount - amount
	outputs = append(outputs, *NewTXOutput(amount, to))
	if change > 0 {
		outputs = append(outputs, *NewTXOutput(change, string(GetAddress(pubKey))))
	}

	tx := Transaction{Vin: inputs, Vout: outputs}
	tx.ID = tx.Hash()
	return &tx, nil
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	if len(tx.Vin) == 1 {
		if tx.Vin[0].OutIdx == -1 && tx.Vin[0].Txid == nil {
			return true
		}
	}
	return false
}

// Equals checks if the given transaction ID matches the ID of tx
func (tx Transaction) Equals(ID []byte) bool {
	txHash := tx.Hash()
	if bytes.Equal(txHash, ID) {
		return true
	}
	return false
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var data bytes.Buffer
	encoder := gob.NewEncoder(&data)
	err := encoder.Encode(Transaction{ID: tx.ID, Vin: tx.Vin, Vout: tx.Vout})
	if err != nil {
		return nil
	}
	return data.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	transaction := *tx
	transaction.ID = []byte{}
	hash := sha256.Sum256(transaction.Serialize())
	return hash[:]
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx Transaction) TrimmedCopy() Transaction {
	var trimmedVin []TXInput
	for _, vin := range tx.Vin {
		trimmedVin = append(trimmedVin, TXInput{Txid: vin.Txid, OutIdx: vin.OutIdx, Signature: nil, PubKey: nil})
	}

	return Transaction{ID: tx.ID, Vin: trimmedVin, Vout: tx.Vout}
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]*Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}

	for _, vin := range tx.Vin {
		if _, ok := prevTXs[hex.EncodeToString(vin.Txid)]; !ok {
			return ErrTxInputNotFound
		}
	}

	trimmedCopy := tx.TrimmedCopy()
	for i, trimVin := range trimmedCopy.Vin {
		trimVin.Signature = nil
		trimVin.PubKey = prevTXs[hex.EncodeToString(trimVin.Txid)].Vout[trimVin.OutIdx].PubKeyHash
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, trimmedCopy.Serialize())
		if err != nil {
			return err
		}
		tx.Vin[i].Signature = append(r.Bytes(), s.Bytes()...)
	}
	return nil
}

// Verify verifies signatures of Transaction inputs
func (tx Transaction) Verify(prevTXs map[string]*Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if _, ok := prevTXs[hex.EncodeToString(vin.Txid)]; !ok {
			return false
		}
	}

	curve := elliptic.P256()
	trimmedCopy := tx.TrimmedCopy()
	for i, vin := range tx.Vin {
		trimmedCopy.Vin[i].Signature = nil
		trimmedCopy.Vin[i].PubKey = prevTXs[hex.EncodeToString(vin.Txid)].Vout[vin.OutIdx].PubKeyHash

		// Recovering sig
		r := big.NewInt(0)
		r.SetBytes(vin.Signature[:len(vin.Signature)/2])
		s := big.NewInt(0)
		s.SetBytes(vin.Signature[len(vin.Signature)/2:])

		// Recovering pubkey
		x := big.NewInt(0)
		x.SetBytes(vin.PubKey[:len(vin.PubKey)/2])
		y := big.NewInt(0)
		y.SetBytes(vin.PubKey[len(vin.PubKey)/2:])

		// Verify
		publicKey := ecdsa.PublicKey{Curve: curve, X: x, Y: y}

		if !ecdsa.Verify(&publicKey, trimmedCopy.Serialize(), r, s) {
			return false
		}
	}
	return true
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x :", tx.ID))

	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       OutIdx:    %d", input.OutIdx))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey: %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       PubKeyHash: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
