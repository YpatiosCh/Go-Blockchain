package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

var (
	MaxNonce = 10000000
)

const targetBits = 12

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork {
	// Setting up how difficult it should be to mine a block:
	// 1. Start with the number 1
	// 2. Move this 1 to the left by (256 - targetBits) spaces
	// For example, if targetBits is 24, we move 1 to the left by (256-24 = 232) spaces
	// This creates a really big number that the hash must be smaller than
	// The more we move to the left, the smaller this number gets, making mining harder
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

// prepareData prepares data that will be hashed
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run performs the proof-of-work mining process to find a valid hash
// Returns:
//   - nonce: The number that made our hash valid
//   - hash: The valid hash that was found
func (pow *ProofOfWork) Run() (int, []byte) {
	// Create variables we'll use in mining:
	var hashInt big.Int // For converting our hash into a number we can compare
	var hash [32]byte   // To store the actual hash we calculate
	nonce := 0          // Our "guess" counter starting at 0

	// Show what data we're trying to mine
	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)

	// Keep trying nonces until we hit the maximum allowed tries
	for nonce < MaxNonce {
		// Combine block data with current nonce
		data := pow.prepareData(nonce)
		// Calculate hash of our data using SHA256
		hash = sha256.Sum256(data)

		// Show the hash we're currently trying (the \r returns to start of line)
		fmt.Printf("\r%x", hash)

		// Convert the hash to a big integer so we can compare it
		hashInt.SetBytes(hash[:])

		// Check if this hash is valid (less than our target)
		// -1 means hashInt is less than target (we found a valid hash!)
		if hashInt.Cmp(pow.target) == -1 {
			// Valid hash found - stop mining
			break
		} else {
			// Hash wasn't valid - try next number
			nonce++
		}
	}

	// Add newlines for clean output formatting
	fmt.Print("\n\n")

	// Return the winning nonce and its hash
	// hash[:] converts our fixed size array to a slice
	return nonce, hash[:]
}

// Validate validates block's PoW
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
