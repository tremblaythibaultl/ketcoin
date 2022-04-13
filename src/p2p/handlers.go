package p2p

import (
	"encoding/json"
	"ketcoin/src/blockchain"
	"log"
	"net"
)

func (n *Node) transactionRequestHandler(JSON []byte) {
	txn := &blockchain.Transaction{}

	err := json.Unmarshal(JSON, txn)
	if err != nil {
		log.Println("Error while decoding transaction")
		log.Println(err)
	}

	if n.validateTransaction(txn) {
		log.Println("Transaction validated, adding it to the mempool")
		n.mutex.Lock()
		n.mempool[txn.Hash[:]] = *txn
		n.mutex.Unlock()
	} else {
		log.Println("Transaction invalid : insufficient balance or invalid signature. Ignoring...")
	}

}

func (n *Node) blockReceptionHandler(JSON []byte) {
	block := &blockchain.Block{}
	err := json.Unmarshal(JSON, block)
	if err != nil {
		log.Println("Error during decoding block")
		log.Println(err)
	}

	n.validateBlock(block)
}

func (n *Node) blockchainReceptionHandler(JSON []byte) {
	bc := &blockchain.Blockchain{}

	err := json.Unmarshal(JSON, bc)
	if err != nil {
		log.Println("Error while decoding blockchain")
		log.Println(err)
	}

	n.validateBlockchain(bc)
}

func (n *Node) blockchainRequestHandler(conn net.Conn) {
	log.Printf("Received bcrq from %s", conn.RemoteAddr())

	log.Println("Starting blockchain transfer...")

	msg, err := n.getBlockchainAsMessage()
	if err != nil {
		log.Println("Error getting blockchain as message")
		log.Println(err)
	}

	n.send(conn, msg)
}
