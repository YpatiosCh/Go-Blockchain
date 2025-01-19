package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

// Global variables defining the proof-of-work parameters
var (
	// maxNonce defines the maximum value the nonce can take.
	// If a solution isn't found after maxNonce iterations,
	// the mining process stops.
	maxNonce = 10000000
)

// targetBits defines the difficulty of mining. The higher this number,
// the easier it is to mine a block. The lower the number, the harder it becomes.
// In Bitcoin, this value is adjusted every 2016 blocks to maintain a consistent
// block generation time of about 10 minutes.
const targetBits = 12

// ProofOfWork represents a proof-of-work system similar to the one used in Bitcoin.
// It ensures that a significant amount of computational work has been invested in
// creating a new block, making it difficult to alter the blockchain.
type ProofOfWork struct {
	block  *Block   // The block to mine
	target *big.Int // The target threshold that the hash must be less than
}

// NewProofOfWork builds and returns a ProofOfWork instance for a given block.
// It calculates the target value based on the targetBits difficulty.
// The target is calculated as: target = 1 << (256 - targetBits)
// This means the hash of the block must be below this target to be valid.
func NewProofOfWork(b *Block) *ProofOfWork {
	// Create a new big integer with value 1
	target := big.NewInt(1)

	// Left shift by (256 - targetBits) positions
	// 256 is used because SHA-256 hash is 256 bits long
	// For example, if targetBits = 12, we shift by 244 positions
	// This creates our target threshold
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

// prepareData combines the block data with the nonce to create
// the data that will be hashed. This implements the core mining algorithm:
// hash(prevHash + transactions + timestamp + targetBits + nonce)
// Parameters:
//   - nonce: The current nonce value being tested
//
// Returns:
//   - []byte: The combined data ready for hashing
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,       // Previous block's hash
			pow.block.HashTransactions(),  // Hash of all transactions in the block
			IntToHex(pow.block.Timestamp), // Block timestamp
			IntToHex(int64(targetBits)),   // Mining difficulty
			IntToHex(int64(nonce)),        // Current nonce value
		},
		[]byte{}, // Separator (empty in this case)
	)

	return data
}

// Run performs the actual proof-of-work computation.
// It continuously hashes the block data with different nonce values
// until it finds a hash that's less than the target.
// Returns:
//   - int: The nonce that produced a valid hash
//   - []byte: The valid hash that was found
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int // Used to store the hash as a big integer for comparison
	var hash [32]byte   // Stores the current hash value
	nonce := 0          // Starting nonce value

	fmt.Printf("Mining a new block")
	for nonce < maxNonce {
		// Prepare the data with the current nonce
		data := pow.prepareData(nonce)

		// Calculate the SHA-256 hash
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash) // Display mining progress

		// Convert hash to big integer for comparison with target
		hashInt.SetBytes(hash[:])

		// Compare hash with target
		// If hash < target, we've found a valid nonce
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++ // Try next nonce value
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate verifies whether a block's proof-of-work is valid.
// It recalculates the hash using the block's nonce and checks if
// it's below the target threshold.
// Returns:
//   - bool: true if the proof-of-work is valid, false otherwise
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	// Recreate the hash using the block's stored nonce
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	// Check if hash is less than target
	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
