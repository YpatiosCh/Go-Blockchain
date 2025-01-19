package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// CLI represents the Command Line Interface for the blockchain application.
// It provides methods to interact with the blockchain through various commands
// such as creating a blockchain, checking balances, sending coins, and viewing
// the blockchain's contents.
type CLI struct{}

// createBlockchain initializes a new blockchain with a genesis block and sends
// the genesis reward to the specified address. This can only be done once - if a
// blockchain already exists, this operation will fail.
// Parameters:
//   - address: The wallet address that will receive the genesis block reward
func (cli *CLI) createBlockchain(address string) {
	bc := CreateBlockchain(address)
	// Ensure we close the database connection when done
	bc.db.Close()
	fmt.Println("Done!")
}

// getBalance calculates and displays the balance for a given wallet address by
// finding all Unspent Transaction Outputs (UTXOs) associated with that address.
// Parameters:
//   - address: The wallet address to check the balance for
func (cli *CLI) getBalance(address string) {
	// Load the existing blockchain
	bc := NewBlockchain(address)
	// Ensure database connection is closed after we're done
	defer bc.db.Close()

	balance := 0
	// Find all unspent transaction outputs for this address
	UTXOs := bc.FindUTXO(address)

	// Sum up the values of all UTXOs
	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

// printUsage displays help information showing all available commands and their
// usage. This is shown when invalid commands are used or when help is requested.
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

// validateArgs checks if any command line arguments were provided.
// If no arguments were given, it prints usage information and exits.
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// printChain displays the entire blockchain, starting from the most recent block
// and moving backwards to the genesis block. For each block, it shows:
// - The previous block's hash
// - The current block's hash
// - Proof of Work validation status
func (cli *CLI) printChain() {
	// Open blockchain without specifying an address since we're just reading
	bc := NewBlockchain("")
	defer bc.db.Close()

	// Create an iterator to move through the blockchain
	bci := bc.Iterator()

	// Iterate through all blocks until we reach the genesis block
	for {
		block := bci.Next()

		// Display block information
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		// Break when we reach the genesis block (it has no previous hash)
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

// send creates a new transaction to transfer coins from one address to another.
// It creates a new transaction, adds it to a new block, and mines the block.
// Parameters:
//   - from: Source wallet address
//   - to: Destination wallet address
//   - amount: Number of coins to transfer
func (cli *CLI) send(from, to string, amount int) {
	// Load the blockchain with the sender's address
	bc := NewBlockchain(from)
	defer bc.db.Close()

	// Create a new UTXO transaction
	tx := NewUTXOTransaction(from, to, amount, bc)
	// Add the transaction to a new block and mine it
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("Success!")
}

// Run is the entry point for the CLI application. It parses command line
// arguments and executes the appropriate command. The supported commands are:
// - getbalance: Check the balance of an address
// - createblockchain: Create a new blockchain
// - printchain: Display all blocks in the chain
// - send: Transfer coins between addresses
func (cli *CLI) Run() {
	cli.validateArgs()

	// Create flag sets for each command
	// flag.ExitOnError means the program will exit if there's an error parsing flags
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	// Define flags for each command
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	// Parse the command from command line arguments
	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	// Execute the appropriate command with its parsed flags
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}
