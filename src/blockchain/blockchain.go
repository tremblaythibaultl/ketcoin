package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
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
		PrevHash:     "",
		Timestamp:    time.Now(),
		Txns:         nil,
		Nonce:        0,
		Reward:       BLOCK_REWARD,
		MinerAddress: a.Address,
	}
	b.Hash = b.ComputeHash()
	// Add genesis block to chain
	bc.Chain = append(bc.Chain, b)
	// Initialize map and add node's account to address => account mapping
	bc.Accounts = make(map[string]*Account)
	bc.Accounts[a.Address] = a
	// Add block reward to node's account
	a.Balance += BLOCK_REWARD
}

func (bc *Blockchain) IsValid() bool {
	for i := 0; i < len(bc.Chain)-1; i++ {
		if bc.Chain[i+1].Index-1 != bc.Chain[i].Index ||
			bc.Chain[i+1].PrevHash != bc.Chain[i].Hash ||
			bc.Chain[i].Hash != bc.Chain[i].ComputeHash() {
			log.Printf("idxs : %d - %d\nh1 : \n%s\n%s\nh2 : \n%s\n%s\n", bc.Chain[i+1].Index-1, bc.Chain[i].Index, bc.Chain[i+1].PrevHash, bc.Chain[i].Hash, bc.Chain[i].Hash, bc.Chain[i].ComputeHash())
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

func (bc *Blockchain) GetStateRoot() string {
	s := ""
	for _, acc := range bc.Accounts {
		s += fmt.Sprintf("%s%d", acc.Address, acc.Balance)
	}
	h := sha256.Sum256([]byte(s))

	return hex.EncodeToString(h[:])
}

// well-defined on a non-zero sized blockchain
func (bc *Blockchain) GetLastIndex() uint64 {
	bc.RLock()
	defer bc.RUnlock()
	return bc.Chain[len(bc.Chain)-1].Index
}

func (bc *Blockchain) GetLastBlock() *Block {
	bc.RLock()
	defer bc.RUnlock()
	return bc.Chain[len(bc.Chain)-1]
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
