package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// CLI represents our command line interface structure
// It holds a pointer to our blockchain so it can interact with it
type CLI struct {
	bc *Blockchain
}

// printUsage shows the user how to use our program
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  addblock -data BLOCK_DATA - add a block to the blockchain")
	fmt.Println("  printchain - print all the blocks of the blockchain")
}

// validateArgs checks if the user provided any arguments
// If not, it shows usage and exits the program
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// addBlock adds a new block with the provided data to the blockchain
func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
	fmt.Println("Success!")
}

// printChain displays all blocks in the blockchain, starting from the newest
func (cli *CLI) printChain() {
	// Create an iterator starting from the newest block
	bci := cli.bc.Iterator()

	// Keep getting blocks until we reach genesis block (which has no previous hash)
	for {
		block := bci.Next()

		// Print block information
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		// Verify the proof of work
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		// If we reach a block with no previous hash (genesis block), stop
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

// Run processes command line arguments and executes appropriate commands
func (cli *CLI) Run() {
	// Check if any arguments were provided
	cli.validateArgs()

	// Create command flags for our two commands
	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	// Add -data flag to addblock command
	addBlockData := addBlockCmd.String("data", "", "Block data")

	// Check which command was used
	switch os.Args[1] {
	case "addblock":
		// Parse addblock command arguments
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		// Parse printchain command arguments
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		// If command not recognized, show usage and exit
		cli.printUsage()
		os.Exit(1)
	}

	// If addblock command was used
	if addBlockCmd.Parsed() {
		// Check if data was provided
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		// Add the block with provided data
		cli.addBlock(*addBlockData)
	}

	// If printchain command was used
	if printChainCmd.Parsed() {
		// Print the entire blockchain
		cli.printChain()
	}
}
