package blockchain

import (
	"fmt"
	"sync"
	"time"
)

const BLOCK_REWARD = 32

type Blockchain struct {
	mutex    sync.RWMutex
	chain    []*block
	accounts map[[32]byte]*Account
}

// TODO : parse chain and create accounts

// Init creates a genesis block.
func (bc *Blockchain) Init(a *Account) {
	b := &block{
		index:        0,
		hash:         "",
		prevHash:     "",
		timestamp:    time.Now().Unix(),
		txns:         nil,
		nonce:        0,
		minerAddress: a.address,
	}
	// Add genesis block to chain
	bc.chain = append(bc.chain, b)
	// Add node's account to address => account mapping
	bc.accounts[a.address] = a
	// Add block reward to node's account
	a.add(BLOCK_REWARD)
}

func (bc *Blockchain) isValid() bool {
	for i := 0; i < len(bc.chain)-1; i++ {
		if bc.chain[i+1].index-1 != bc.chain[i].index ||
			bc.chain[i+1].prevHash != bc.chain[i].hash ||
			bc.chain[i].hash != bc.chain[i].computeHash() {
			return false
		}
	}
	return true
}

func (bc *Blockchain) addBlock(b *block) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	bc.chain = append(bc.chain, b)
}

func (bc *Blockchain) replaceChain(other *Blockchain) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	bc.chain = other.chain
}

func (bc *Blockchain) print() {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	s := "["
	for _, block := range bc.chain {
		s += block.prettyPrint() + ","
	}
	s += "]"
	fmt.Println(s)
}
