package main

import (
	"log"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db" // The name of the database file
const blocksBucket = "blocks"  // The name of the bucket where we'll store our blocks (think of it like a table in SQL)

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// AddBlock adds a new block to the blockchain with the provided data
func (bc *Blockchain) AddBlock(data string) {
	// Variable to store the hash of the previous (last) block
	var lastHash []byte

	// STEP 1: GET THE LAST BLOCK'S HASH
	// Start a read-only transaction (View) to get the last hash
	err := bc.db.View(func(tx *bolt.Tx) error {
		// Get our bucket (table) of blocks
		b := tx.Bucket([]byte(blocksBucket))
		// Get the hash stored under "l" (the last block's hash)
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// STEP 2: CREATE NEW BLOCK
	// Create a new block with our data and the hash of the last block
	newBlock := NewBlock(data, lastHash)

	// STEP 3: SAVE THE NEW BLOCK
	// Start a write transaction (Update) to save our new block
	// Think of it like adding a new page to our book
	err = bc.db.Update(func(tx *bolt.Tx) error {
		// Get our bucket again
		b := tx.Bucket([]byte(blocksBucket))

		// Save the new block in the database:
		// Key: new block's hash
		// Value: serialized block data
		err = b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		// Update our "last block" bookmark to point to this new block
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		// Update our tip (bookmark) in memory to point to the new block
		bc.tip = newBlock.Hash

		return nil
	})
}

// Iterator creates a new BlockchainIterator that will let us go through all blocks
// Starting from the newest block and moving backwards to the genesis block
func (bc *Blockchain) Iterator() *BlockchainIterator {
	// Create a new iterator that starts at:
	// - bc.tip: hash of the newest/latest block (our starting point)
	// - bc.db: reference to our database so we can read blocks
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// Next gets the next block in the blockchain (moving backwards from newest to oldest)
func (i *BlockchainIterator) Next() *Block {
	// Variable to store the block we'll return
	var block *Block

	// Start a read-only database transaction
	err := i.db.View(func(tx *bolt.Tx) error {
		// Get our bucket of blocks
		b := tx.Bucket([]byte(blocksBucket))

		// Get the serialized (encoded) block using current hash
		encodedBlock := b.Get(i.currentHash)

		// Convert the encoded data back into a Block structure
		block = DeserializeBlock(encodedBlock)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// Update currentHash to point to the previous block
	// This is how we move backwards through the chain
	i.currentHash = block.PrevBlockHash

	return block
}

// NewBlockchain creates a new Blockchain with Genesis Block if it doesn't exist
// or loads an existing blockchain from the database
func NewBlockchain() *Blockchain {
	// tip will store the hash of the last block in the chain
	var tip []byte

	// Open the database file. If it doesn't exist, it will be created
	// 0600 means read-write permissions for the owner only
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	// Start a database transaction where we can make changes
	err = db.Update(func(tx *bolt.Tx) error {
		// Try to get the bucket (think of it like a table) where we store blocks
		b := tx.Bucket([]byte(blocksBucket))

		// If the bucket doesn't exist (first time running the program)
		if b == nil {
			// Create the first (genesis) block
			genesis := NewGenesisBlock()

			// Create a new bucket to store our blocks
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}
			// Store the genesis block in the database:
			// Key: the block's hash
			// Value: the serialized block data
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			// Store the hash of the last block with key "l"
			// This helps us know where the chain ends
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			// Keep track of the last block's hash
			tip = genesis.Hash
		} else {
			// If bucket exists, just get the last block's hash
			tip = b.Get([]byte("l"))
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	// Create and return a new Blockchain with:
	// - tip: hash of the last block
	// - db: reference to the database
	bc := Blockchain{tip, db}
	return &bc
}
