package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

// Database configuration constants
const dbFile = "blockchain.db" // The file where the blockchain data is stored
const blocksBucket = "blocks"  // The bucket (similar to a table) name in BoltDB
// The message included in the genesis block, referencing The Times headline
// This is the same message that was included in Bitcoin's genesis block
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// Blockchain represents a chain of blocks stored in a BoltDB database.
// It maintains a reference to the last block (tip) and the database connection.
type Blockchain struct {
	tip []byte   // Hash of the last block in the chain
	db  *bolt.DB // Database connection
}

// BlockchainIterator provides functionality to iterate over blockchain blocks
// from newest to oldest (back to genesis block)
type BlockchainIterator struct {
	currentHash []byte   // Hash of the current block
	db          *bolt.DB // Database connection
}

// MineBlock creates a new block with the provided transactions and adds it to the chain.
// This simulates the mining process in a real blockchain network.
// Parameters:
//   - transactions: Array of transactions to include in the new block
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	// Retrieve the last block's hash from the database
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l")) // 'l' key stores the last block's hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// Create new block with the transactions
	newBlock := NewBlock(transactions, lastHash)

	// Store the new block in the database
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		// Store the serialized block
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		// Update the 'l' key to point to our new block
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		// Update the tip
		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// FindUnspentTransactions scans the blockchain and returns all unspent transactions
// for a given address. This is a key function for the UTXO (Unspent Transaction Output) model.
// Parameters:
//   - address: The address to find unspent transactions for
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int) // Maps transaction IDs to spent output indices
	bci := bc.Iterator()

	// Iterate through all blocks
	for {
		block := bci.Next()

		// Look at each transaction in the block
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			// Check each output in the transaction
		Outputs:
			for outIdx, out := range tx.Vout {
				// Skip if output was already spent
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				// If output can be unlocked by the provided address, it's unspent
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			// If not a coinbase transaction, mark its inputs as spent
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		// Break when we reach the genesis block
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// FindUTXO finds all unspent transaction outputs for an address
// and returns them. This is used to calculate account balance.
// Parameters:
//   - address: The address to find UTXOs for
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	// Collect all outputs that can be unlocked by the address
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// FindSpendableOutputs finds enough unspent outputs to cover the requested amount.
// This is used when creating new transactions, to find outputs to use as inputs.
// Parameters:
//   - address: The address to find spendable outputs for
//   - amount: The amount needed
//
// Returns:
//   - accumulated: The total amount found
//   - unspentOutputs: Map of transaction IDs to output indices
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

// Iterator creates and returns a BlockchainIterator instance
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db}
}

// Next returns the next block in the chain.
// Blocks are returned in reverse order (newest to oldest)
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	// Read the block from database
	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// Move to the previous block
	i.currentHash = block.PrevBlockHash

	return block
}

// dbExists checks if the blockchain database file exists
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// NewBlockchain creates a new Blockchain instance, loading an existing chain from the database.
// Parameters:
//   - address: The address to work with (not used in basic implementation)
func NewBlockchain(address string) *Blockchain {
	if !dbExists() {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	// Get the last block hash
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	return &bc
}

// CreateBlockchain creates a new blockchain DB with a genesis block
// Parameters:
//   - address: The address to send the genesis block reward to
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	// Initialize the blockchain with genesis block
	err = db.Update(func(tx *bolt.Tx) error {
		// Create the coinbase transaction for genesis block
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		// Create the blocks bucket
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		// Store the genesis block
		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		// Update the 'l' key to point to genesis block
		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	return &bc
}
