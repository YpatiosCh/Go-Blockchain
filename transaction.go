package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// subsidy is the amount of reward given for mining a new block.
// In Bitcoin, this value is halved approximately every 4 years.
// Starting at 50 BTC, then 25 BTC, 12.5 BTC, and so on.
const subsidy = 10

// Transaction represents a blockchain transaction, similar to Bitcoin's structure.
// It contains inputs (references to previous outputs) and outputs (new coins).
// The transaction ID is a hash of the entire transaction data.
type Transaction struct {
	ID   []byte     // Unique identifier of the transaction (hash of its contents)
	Vin  []TXInput  // Array of transaction inputs (money being spent)
	Vout []TXOutput // Array of transaction outputs (money being created/transferred)
}

// IsCoinbase checks whether the transaction is a coinbase transaction.
// Coinbase transactions are special transactions that create new coins and are
// included as the first transaction in each block as a mining reward.
// They have specific characteristics:
// 1. Only one input
// 2. Input txID is empty (no previous transaction)
// 3. Input Vout is -1 (no previous output index)
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// SetID calculates and sets the transaction ID.
// The ID is a SHA-256 hash of the entire transaction data (inputs and outputs)
// encoded using GOB encoding (Go's binary format).
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	// Create a new GOB encoder and encode the transaction
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	// Calculate SHA-256 hash of the encoded transaction
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// TXInput represents a transaction input.
// In a blockchain, inputs are references to previous transaction outputs
// that are being spent in the current transaction.
type TXInput struct {
	Txid      []byte // The ID of the transaction containing the output being referenced
	Vout      int    // The index of the output in the referenced transaction
	ScriptSig string // The script that provides data to be validated against the output's ScriptPubKey
}

// TXOutput represents a transaction output.
// Outputs are new coins created by the transaction, which can later
// be referenced as inputs in new transactions (when being spent).
type TXOutput struct {
	Value        int    // The amount of coins
	ScriptPubKey string // The script that specifies spending conditions (usually contains the owner's address)
}

// CanUnlockOutputWith checks if the provided data can unlock this input.
// This is a simplified version of Bitcoin's Script system.
// In real Bitcoin, this would involve executing Script code.
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

// CanBeUnlockedWith checks if the provided data can unlock this output.
// This is also a simplified version of Bitcoin's Script system.
// In real Bitcoin, this would involve executing Script code.
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

// NewCoinbaseTX creates a new coinbase transaction.
// Coinbase transactions are special transactions that create new coins.
// They're used to reward miners for mining blocks.
// Parameters:
//   - to: The address that will receive the mining reward
//   - data: Optional data to include in the transaction (like a message)
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	// Create input: empty txID, vout = -1, and data as ScriptSig
	txin := TXInput{[]byte{}, -1, data}
	// Create output: value = mining reward, ScriptPubKey = recipient's address
	txout := TXOutput{subsidy, to}
	// Create and return the transaction
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

// NewUTXOTransaction creates a new transaction transferring value between addresses.
// This implements the UTXO (Unspent Transaction Output) model used by Bitcoin.
// Parameters:
//   - from: Sender's address
//   - to: Recipient's address
//   - amount: Amount to send
//   - bc: Pointer to the blockchain to verify and find UTXOs
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	// Find and verify sufficient funds in the blockchain
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs by referencing previous outputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		// Create an input for each output we're spending
		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	// First output is the payment to the recipient
	outputs = append(outputs, TXOutput{amount, to})

	// If there are leftover funds, send them back to sender as change
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}

	// Create, set ID, and return the transaction
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}
