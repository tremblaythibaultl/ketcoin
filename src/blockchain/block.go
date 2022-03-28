package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type transaction struct {
	Sender   string
	Receiver string
	Data     string
}

type mempool struct {
	pendingTxns []transaction
}

type Block struct {
	Index        uint64
	Hash         string
	PrevHash     string
	Timestamp    int64
	Txns         []transaction
	Nonce        int
	MinerAddress string
}

func (b *Block) prettyPrint() string {
	var s = fmt.Sprintf("Block %d@%d hash %s\n", b.Index, b.Timestamp, b.Hash)
	for i := 0; i < len(b.Txns); i++ {
		s += fmt.Sprintf("txn %d from %s to %s : %s\n", i, b.Txns[i].Sender, b.Txns[i].Receiver, b.Txns[i].Data)
	}
	return s
}

func (b *Block) toString() string {
	return fmt.Sprintf("%d%s%d", b.Index, b.PrevHash, b.Nonce)
}

// ComputeHash computes the Block's hash. Used in mining.
func (b *Block) ComputeHash() string {
	h := sha256.Sum256([]byte(b.toString()))
	return hex.EncodeToString(h[:])
}

func (b *Block) GetIndex() uint64 {
	return b.Index
}

func (b *Block) mine() bool {
	for invalid := true; invalid; invalid = b.ComputeHash()[0:1] != "0" { //hardcoded difficulty for testing purposes
		fmt.Println(b.ComputeHash())
		b.Nonce++
	}
	return false
}
