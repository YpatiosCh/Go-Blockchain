package main

type BlockChain struct {
	blocks []*Block
}

// AddBlock adds a new block to the blockchain
func (bc *BlockChain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]   // Get the last block
	newBlock := NewBlock(data, prevBlock.Hash) // Create a new block
	bc.blocks = append(bc.blocks, newBlock)    // Append the new block to the blockchain
}

// NewBlockChain creates a new blockchain with the genesis block
func NewBlockChain() *BlockChain {
	return &BlockChain{[]*Block{NewGenesisBlock()}}
}
