package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

// Block represents a block in the blockchain.
// Each block contains:
// - Timestamp: When the block was created
// - Transactions: List of transactions included in the block
// - PrevBlockHash: Hash of the previous block (forms the chain)
// - Hash: Hash of the current block
// - Nonce: Number used in the proof-of-work algorithm
type Block struct {
	Timestamp     int64          // Unix timestamp when the block was created
	Transactions  []*Transaction // List of transactions included in this block
	PrevBlockHash []byte         // Reference to previous block's hash
	Hash          []byte         // This block's hash (computed based on block contents)
	Nonce         int            // Nonce used to generate a hash meeting the mining difficulty requirements
}

// Serialize converts the Block struct into a byte array.
// This is necessary for storing the block in the database.
// The function uses Go's encoding/gob package for binary serialization.
// Returns:
//   - []byte: Serialized block data
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	// Create a new GOB encoder writing to our buffer
	encoder := gob.NewEncoder(&result)

	// Encode the entire block structure
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// HashTransactions creates a hash of all transactions in the block.
// This hash is used as part of the block's header and ensures that
// transaction data cannot be tampered with.
// The function:
// 1. Collects all transaction IDs
// 2. Concatenates them
// 3. Creates a SHA-256 hash of the concatenated data
// Returns:
//   - []byte: Hash of all transactions
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	// Collect all transaction IDs
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	// Create a single hash of all transaction hashes
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

// NewBlock creates and returns a new Block.
// This function:
// 1. Creates a basic block with the provided data
// 2. Performs proof-of-work to generate valid hash
// 3. Sets the computed hash and nonce
// Parameters:
//   - transactions: List of transactions to include in the block
//   - prevBlockHash: Hash of the previous block in the chain
//
// Returns:
//   - *Block: Newly created and mined block
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	// Create basic block structure with current timestamp
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}

	// Create a proof-of-work instance for this block
	pow := NewProofOfWork(block)
	// Run mining process to find valid hash and nonce
	nonce, hash := pow.Run()

	// Set the computed values
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns the genesis block.
// The genesis block is the first block in the blockchain.
// It's special because it has no previous block hash.
// Parameters:
//   - coinbase: The coinbase transaction for the genesis block
//
// Returns:
//   - *Block: The genesis block
func NewGenesisBlock(coinbase *Transaction) *Block {
	// Create new block with no previous hash (empty byte array)
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// DeserializeBlock converts a byte array back into a Block struct.
// This is used when reading blocks from the database.
// Parameters:
//   - d: Serialized block data
//
// Returns:
//   - *Block: Deserialized block structure
func DeserializeBlock(d []byte) *Block {
	var block Block

	// Create a GOB decoder reading from our bytes
	decoder := gob.NewDecoder(bytes.NewReader(d))
	// Decode bytes into a Block structure
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
