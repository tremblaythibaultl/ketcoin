package blockchain

import (
	"fmt"
	"sync"
	"time"
)

const BLOCK_REWARD = 32

type Blockchain struct {
	mutex    sync.RWMutex
	Chain    []*Block
	Accounts map[string]*Account
}

// Init creates a genesis block.
func (bc *Blockchain) Init(a *Account) {
	b := &Block{
		Index:        0,
		Hash:         "",
		PrevHash:     "",
		Timestamp:    time.Now().Unix(),
		Txns:         nil,
		Nonce:        0,
		MinerAddress: a.Address,
	}
	// Add genesis block to chain
	bc.Chain = append(bc.Chain, b)
	// Initialize map and add node's account to address => account mapping
	bc.Accounts = make(map[string]*Account)
	bc.Accounts[a.Address] = a
	// Add block reward to node's account
	a.add(BLOCK_REWARD)
}

func (bc *Blockchain) IsValid() bool {
	for i := 0; i < len(bc.Chain)-1; i++ {
		if bc.Chain[i+1].Index-1 != bc.Chain[i].Index ||
			bc.Chain[i+1].PrevHash != bc.Chain[i].Hash ||
			bc.Chain[i].Hash != bc.Chain[i].ComputeHash() {
			return false
		}
	}
	return true
}

func (bc *Blockchain) AddBlock(b *Block) {
	bc.Lock()
	defer bc.Unlock()
	bc.Chain = append(bc.Chain, b)
}

func (bc *Blockchain) ReplaceChain(other *Blockchain) {
	bc.Lock()
	defer bc.Unlock()
	bc.Chain = other.Chain
	bc.Accounts = other.Accounts //will this import accounts too?
}

func (bc *Blockchain) GetLastIndex() uint64 {
	bc.RLock()
	defer bc.RUnlock()
	return bc.Chain[len(bc.Chain)-1].Index
}

//Encapsulation mutex function to avoid blockchain encoding errors
func (bc *Blockchain) Lock() {
	bc.mutex.Lock()
}

func (bc *Blockchain) Unlock() {
	bc.mutex.Unlock()
}

func (bc *Blockchain) RLock() {
	bc.mutex.RLock()
}

func (bc *Blockchain) RUnlock() {
	bc.mutex.RUnlock()
}

func (bc *Blockchain) print() {
	bc.RLock()
	defer bc.RUnlock()
	s := "["
	for _, block := range bc.Chain {
		s += block.prettyPrint() + ","
	}
	s += "]"
	fmt.Println(s)
}
