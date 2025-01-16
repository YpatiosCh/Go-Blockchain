package main

import "fmt"

func main() {
	bc := NewBlockChain() // Create a new blockchain

	bc.AddBlock("Send 1 BTC to Ivan")      // Add a new block to the blockchain
	bc.AddBlock("Send 2 more BTC to Ivan") // Add a new block to the blockchain

	for _, block := range bc.blocks {
		fmt.Printf("Prev.Hash: %x\n", block.PrevBlockHash) // Print the previous block's hash
		fmt.Printf("Data: %s\n", block.Data)               // Print the block's data
		fmt.Printf("Hash: %x\n", block.Hash)               // Print the block's hash
		fmt.Println()
	}
}
