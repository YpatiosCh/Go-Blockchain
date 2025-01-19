# Simple Blockchain Implementation in Go

This project implements a basic blockchain with key features similar to Bitcoin, including proof-of-work mining, transactions with the UTXO model, and persistent storage. It's built as an educational tool to understand how blockchains work. Inspired by Jeiwan's work @ https://jeiwan.net/ 

## Table of Contents
- [Features](#features)
- [Technology Stack](#technology-stack)
- [Core Components](#core-components)
- [How It Works](#how-it-works)
- [Installation](#installation)
- [Usage](#usage)
- [Technical Details](#technical-details)

## Features

- Proof of Work (PoW) mining system
- UTXO (Unspent Transaction Output) model
- Persistent storage using BoltDB
- Command-line interface
- Transaction creation and validation
- Balance querying
- Blockchain exploration

## Technology Stack

- **Go** (Golang) - Main programming language
- **BoltDB** - Key-value store for blockchain data
- **crypto/sha256** - For cryptographic hashing
- **encoding/gob** - For Go binary serialization
- **flag** package - For command-line argument parsing

## Core Components

### 1. Block Structure
```go
type Block struct {
    Timestamp     int64
    Transactions  []*Transaction
    PrevBlockHash []byte
    Hash          []byte
    Nonce         int
}
```
- Stores transaction data and metadata
- Links to previous block through PrevBlockHash
- Includes proof-of-work nonce

### 2. Transaction System
```go
type Transaction struct {
    ID   []byte
    Vin  []TXInput
    Vout []TXOutput
}
```
- Implements UTXO model
- Tracks inputs (spent coins) and outputs (new coins)
- Supports coinbase transactions (mining rewards)

### 3. Blockchain
```go
type Blockchain struct {
    tip []byte
    db  *bolt.DB
}
```
- Manages the chain of blocks
- Handles persistent storage
- Provides iteration capabilities

## How It Works

### 1. Mining Process
1. New transactions are collected
2. Block is created with these transactions
3. Proof of Work algorithm runs to find valid nonce
4. Block is added to chain when valid hash is found

### 2. Transaction Flow
1. User initiates a transaction
2. System finds unspent outputs (UTXO) for the sender
3. New transaction is created with inputs and outputs
4. Transaction is included in a new block
5. Block is mined and added to chain

### 3. Data Storage
- Uses BoltDB as key-value store
- Blocks are serialized using Go's gob encoding
- Each block is stored with its hash as the key
- Special key 'l' tracks the latest block hash

## Installation

1. Install Go (1.13 or later)
2. Install BoltDB:
```bash
go get github.com/boltdb/bolt
```
3. Clone this repository:
```bash
git clone https://github.com/YpatiosCh/Go-Blockchain.git
cd go-blockchain
```
4. Build the project:
```bash
go build
```

## Usage

### Create a New Blockchain
```bash
./go-blockchain createblockchain -address {PERSON}
```
Creates a new blockchain and sends genesis reward to {PERSON}

### Get Balance
```bash
./go-blockchain getbalance -address {PERSON}
```
Shows the balance for the specified address

### Send Coins
```bash
./go-blockchain send -from {PERSON} -to {PERSON} -amount AMOUNT
```
Sends AMOUNT of coins from {PERSON} address to {PERSON} address

### Print Chain
```bash
./go-blockchain printchain
```
Prints all blocks in the blockchain

## Technical Details

### Proof of Work
- Uses SHA-256 hashing
- Target difficulty: 12 bits (adjustable)
- Nonce limit: 10000000
- Hash must be below target to be valid

### Transaction Verification
1. Input validation
   - Checks if referenced outputs exist
   - Verifies ownership (simple address matching)
2. Output validation
   - Ensures total output <= total input
   - Validates output structure

### UTXO Management
1. Tracks all unspent transactions
2. Uses memory map for quick lookup
3. Updates after each new block
4. Handles address-based queries

### Database Structure
- Single bucket 'blocks' stores all data
- Block hash → Serialized block data
- Special key 'l' → Latest block hash
- Genesis block includes special coinbase message

### Security Features
- Immutable block history
- Cryptographic linking of blocks
- Transaction validation
- Proof of Work consensus

## Design Decisions

### 1. Choice of BoltDB
- ACID compliant
- Simple key-value structure
- Perfect for blockchain's append-only nature
- Fast read performance

### 2. Transaction Model
- Based on Bitcoin's UTXO model
- Simpler than account-based model
- Better privacy characteristics
- Natural support for parallel processing

### 3. Proof of Work
- Simple to implement
- Easy to verify
- Resource-intensive to compute
- Adjustable difficulty

## Limitations

1. **Simplified Security**: No public/private key cryptography
2. **No Networking**: Single node operation only
3. **Basic Consensus**: No fork resolution
4. **Memory Usage**: Full chain loaded for some operations
5. **Fixed Difficulty**: No dynamic difficulty adjustment

## Future Improvements

1. Add public key cryptography
2. Implement networking layer
3. Add dynamic difficulty adjustment
4. Improve UTXO caching
5. Add support for smart contracts
6. Implement Merkle trees for efficient verification

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.