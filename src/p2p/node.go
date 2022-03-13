package p2p

import (
	"fmt"
	"ketcoin/src/blockchain"
	"log"
	"net"
	"sync"
)

type Node struct {
	listenPort  uint16
	address     string
	connections chan net.Conn
	peers       sync.Map
	blockchain  *blockchain.Blockchain
	account     *blockchain.Account
}

func MakeNode(port uint16) *Node {
	return &Node{
		listenPort: port,
		blockchain: new(blockchain.Blockchain),
	}
}

func (n *Node) Init(target *string) {
	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", n.listenPort))
	if err != nil {
		log.Println("Error starting listener")
		log.Println(err)
	}

	if *target != "" {
		//TODO : ADD PEER, DOWNLOAD AND VALIDATE BCHAIN
	} else { // initialize blockchain
		n.blockchain.Init(n.account)
	}
	n.address = listener.Addr().String()
	log.Printf("Listening on address : %s", n.address)
	n.connections = make(chan net.Conn)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Error while accepting connection")
			}
			n.connections <- conn
		}
	}()
}
