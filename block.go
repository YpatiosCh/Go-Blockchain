package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

// Block keeps block headers
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// NewBlock creates and returns Block
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// Serialize converts our Block structure into a slice of bytes
// so it can be saved in the database
func (b *Block) Serialize() []byte {
	// Create an empty container (buffer) that will hold our bytes
	// Think of it like an empty box that can grow or shrink to fit what we put in it
	var result bytes.Buffer

	// Create a new encoder that will help us pack our Block data into bytes
	// Think of it like getting a machine that knows how to pack things efficiently
	encoder := gob.NewEncoder(&result)

	// Try to convert (encode) our Block structure into bytes and put them in our buffer
	// This takes ALL the Block data (timestamp, hash, data, etc.) and turns it into bytes
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	// Return all our packed bytes
	return result.Bytes()
}

// DeserializeBlock takes bytes from the database and converts them back into a Block structure
// It's like unpacking a box to reconstruct exactly what was in it
func DeserializeBlock(d []byte) *Block {
	// Create an empty Block structure that we'll fill with our unpacked data
	var block Block

	// Create a new decoder that will help us unpack the bytes back into a Block structure
	// bytes.NewReader(d) creates a new reader that reads from our byte slice
	decoder := gob.NewDecoder(bytes.NewReader(d))

	// Try to convert (decode) our bytes back into the Block structure
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
