package p2p

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"ketcoin/src/blockchain"
	"log"
	"net"
)

func (n *Node) blockReceptionHandler(g string) {
	buf := new(bytes.Buffer)
	byteData, err := base64.StdEncoding.DecodeString(g)
	if err != nil {
		log.Println("Error during decoding")
		log.Println(err)
	}
	buf.Write(byteData)
	b := blockchain.Block{}
	err = gob.NewDecoder(buf).Decode(&b)
	if err != nil {
		log.Println("Error while ungobbing block")
		log.Println(err)
	}

	n.validateBlock(&b)
}

func (n *Node) blockchainReceptionHandler(g string) {
	buf := new(bytes.Buffer)
	byteData, err := base64.StdEncoding.DecodeString(g)
	if err != nil {
		log.Println("Error during decoding")
		log.Println(err)
	}
	buf.Write(byteData)
	bc := blockchain.Blockchain{}
	err = gob.NewDecoder(buf).Decode(&bc)
	if err != nil {
		log.Println("Error while ungobbing blockchain")
		log.Println(err)
	}

	n.validateBlockchain(&bc)
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
