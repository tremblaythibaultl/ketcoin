package p2p

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"ketcoin/src/blockchain"
	"ketcoin/src/crypto"
	"log"
	"net"
	"sync"
	"time"
)

const DIALTIMEOUT = time.Second * 5

type Node struct {
	listenPort  uint16
	address     string
	connections chan net.Conn
	peers       sync.Map
	blockchain  *blockchain.Blockchain
	account     *blockchain.Account
	sigTree     *crypto.MerkleSigTree
}

type Message struct {
	Rpc string
	Gob string
}

func MakeNode(port uint16) *Node {
	return &Node{
		listenPort: port,
		blockchain: new(blockchain.Blockchain),
	}
}

func (n *Node) Start() {
	for {
		go n.handle(<-n.connections)
	}
}

func (n *Node) handle(conn net.Conn) {
	defer conn.Close()

	m := new(Message)
	err := json.NewDecoder(conn).Decode(m)
	if err != nil {
		log.Printf("Error decoding message from %s", conn.RemoteAddr())
		log.Println(err)
	}

	log.Printf("Received RPC : %s\nwith data : %s\nfrom : %s", m.Rpc, m.Gob, conn.RemoteAddr())

	switch m.Rpc {
	case "blockchainrequest":
		n.blockchainRequestHandler(conn)
	case "blockchainreception":
		n.blockchainReceptionHandler(m.Gob)
	case "blockreception":
		n.blockReceptionHandler(m.Gob)
	default:
		log.Printf("Remote procedure call %s does not exist on this client, ignoring...", m.Rpc)
	}
}

func (n *Node) send(conn net.Conn, m *Message) {
	err := json.NewEncoder(conn).Encode(m)
	if err != nil {
		log.Printf("Error enconding message going to %s", conn.RemoteAddr())
		log.Println(err)
	}
}

func (n *Node) validateBlock(b *blockchain.Block) {
	if b.ComputeHash()[0:1] == "0" && b.GetIndex() == n.blockchain.GetLastIndex()+1 {
		log.Println("Received block validated, appending to chain")
		n.blockchain.AddBlock(b)
	} else {
		log.Println("Received block not valid, ignoring...")
	}
}

func (n *Node) validateBlockchain(bc *blockchain.Blockchain) {
	if bc.GetLastIndex() > n.blockchain.GetLastIndex() {
		log.Println("Received blockchain has higher index, validating blockchain...")
		if bc.IsValid() {
			log.Println("Received blockchain is longer and valid, replacing blockchain...")
			n.blockchain.ReplaceChain(bc)
		}
	}
}

func (n *Node) getBlockchainAsMessage() (*Message, error) {
	n.blockchain.Lock()
	defer n.blockchain.Unlock()
	var blockchainData bytes.Buffer
	err := gob.NewEncoder(&blockchainData).Encode(n.blockchain)
	if err != nil {
		log.Println("Error while gobbing blockchain")
		log.Println(err)
	} else {
		return &Message{
			Rpc: "blockchainreception",
			Gob: blockchainData.String(),
		}, nil
	}
	return nil, err
}

func (n *Node) requestBlockchain(conn net.Conn) {
	msg := &Message{
		Rpc: "blockchainrequest",
		Gob: "",
	}
	n.send(conn, msg)
}

func (n *Node) Init(target *string) {
	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", n.listenPort))
	if err != nil {
		log.Println("Error starting listener")
		log.Println(err)
	}
	n.sigTree = crypto.NewMSS()
	n.account = &blockchain.Account{
		Address: string(n.sigTree.GetPublicKey()),
		Balance: 0,
	}
	if *target != "" {
		//TODO : ADD PEER, DOWNLOAD AND VALIDATE BCHAIN
		log.Println("Trying to add peer %s", target)
		conn, err := net.DialTimeout("tcp", *target, DIALTIMEOUT)
		if err != nil {
			log.Println("Error dialing target %s", *target)
			log.Println(err)
		}
		n.peers.Store(*target, true)
		log.Print("Added peer %s", *target)
		n.requestBlockchain(conn)
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
