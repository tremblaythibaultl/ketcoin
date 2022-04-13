package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"ketcoin/src/crypto"
	"time"
)

type Transaction struct {
	Sender    string
	Receiver  string
	Amount    uint64
	Timestamp time.Time
	Signature *crypto.MssSignature
	Hash      string
}

type Block struct {
	Index        uint64
	Hash         string
	PrevHash     string
	Timestamp    time.Time
	Txns         []Transaction
	StateRoot    string
	Nonce        int
	Reward       int
	MinerAddress string
}

func (b *Block) prettyPrint() string {
	s := fmt.Sprintf("Block %d@%d hash %s\n", b.Index, b.Timestamp, b.Hash)
	for i := 0; i < len(b.Txns); i++ {
		s += fmt.Sprintf("txn %d from %s to %s : %s\n", i, b.Txns[i].Sender, b.Txns[i].Receiver, b.Txns[i].Amount)
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

func (t *Transaction) ComputeHash() string {
	s := fmt.Sprintf("%s%s%d%d%d", t.Sender, t.Receiver, t.Amount, t.Timestamp.Unix())
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
