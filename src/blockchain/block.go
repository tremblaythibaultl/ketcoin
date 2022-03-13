package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type transaction struct {
	sender   string
	receiver string
	data     string
}

type block struct {
	index        uint64
	hash         string
	prevHash     string
	timestamp    int64
	txns         []transaction
	nonce        int
	minerAddress [32]byte
}

func (b *block) prettyPrint() string {
	var s = fmt.Sprintf("Block %d@%d hash %s\n", b.index, b.timestamp, b.hash)
	for i := 0; i < len(b.txns); i++ {
		s += fmt.Sprintf("txn %d from %s to %s : %s\n", i, b.txns[i].sender, b.txns[i].receiver, b.txns[i].data)
	}
	return s
}

func (b *block) toString() string {
	return fmt.Sprintf("%d%s%d", b.index, b.prevHash, b.nonce)
}

// computeHash computes the block's hash. Used in mining.
func (b *block) computeHash() string {
	h := sha256.Sum256([]byte(b.toString()))
	return hex.EncodeToString(h[:])
}
