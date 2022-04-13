package p2p

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"ketcoin/src/blockchain"
	"ketcoin/src/crypto"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const DIALTIMEOUT = time.Second * 5

type Node struct {
	mutex       sync.RWMutex
	listenPort  uint16
	address     string
	connections chan net.Conn
	peers       sync.Map
	blockchain  *blockchain.Blockchain
	account     *blockchain.Account
	sigTree     *crypto.MerkleSigTree
	mempool     map[string]blockchain.Transaction
}

type Message struct {
	Rpc  string
	JSON []byte
}

func MakeNode(port uint16) *Node {
	return &Node{
		listenPort: port,
		blockchain: new(blockchain.Blockchain),
	}
}

func (n *Node) Start() {
	go n.mine()
	for {
		go n.handle(<-n.connections)
	}
}

func (n *Node) validateTransaction(t *blockchain.Transaction) bool {
	var valid bool
	if acc, exists := n.blockchain.Accounts[t.Sender]; exists {
		valid = acc.Balance-t.Amount >= 0
		if acc.Balance-t.Amount < 0 {
			log.Printf("Invalid transaction ; insufficient balance")
		}
	} else {
		acc = &blockchain.Account{
			Address: t.Sender,
			Balance: 0,
		}
		n.blockchain.Accounts[t.Sender] = acc
		valid = false
	}
	if valid {
		addr := make([]byte, 32)
		hex.Decode(addr, []byte(t.Sender))
		hash := make([]byte, 32)
		hex.Decode(hash, []byte(t.Hash))

		valid = crypto.Verify(t.Signature, *(*[32]byte)(addr), *(*[32]byte)(hash))
		if !valid {
			log.Printf("Invalid transaction ; incorrect signature")
		}
	}
	log.Println(n.blockchain.Accounts[t.Sender])
	return valid
}

func (n *Node) getTransactionList() []blockchain.Transaction {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	txns := make([]blockchain.Transaction, 0, len(n.mempool))
	for _, val := range n.mempool {
		txns = append(txns, val)
	}
	return txns
}

func (n *Node) generateBlock() *blockchain.Block {
	n.blockchain.RLock()
	defer n.blockchain.RUnlock()
	b := &blockchain.Block{
		Index:        n.blockchain.GetLastBlock().Index + 1,
		PrevHash:     n.blockchain.GetLastBlock().Hash,
		Timestamp:    time.Now(),
		Txns:         n.getTransactionList(),
		Nonce:        0,
		Reward:       blockchain.BLOCK_REWARD,
		MinerAddress: n.account.Address,
	}

	return b
}

func (n *Node) broadcastBlock(b *blockchain.Block) {
	blockData, err := json.Marshal(b)
	if err != nil {
		log.Println("Error while encoding block")
		log.Println(err)
	} else {
		m := &Message{
			Rpc:  "blockreception",
			JSON: blockData,
		}
		n.peers.Range(func(k, v interface{}) bool {
			conn := k.(net.Conn)
			isValid := v.(bool)
			if isValid {
				n.send(conn, m)
			}
			return true
		})
	}

}

func (n *Node) printState() {
	for _, acc := range n.blockchain.Accounts {
		log.Println(acc)
	}
}

func (n *Node) mine() {
	log.Println("currently mining")
	for {
		sleepCounter := 0
		for len(n.mempool) == 0 {
			time.Sleep(time.Second)
			log.Println("mempool empty")
			sleepCounter++
			if (sleepCounter % 5) == 0 {
				n.printState()
			}
		}

		txnNb := len(n.mempool)
		var b *blockchain.Block
		for {
			b = n.generateBlock()
			for txnNb >= len(n.mempool) && b.Index > n.blockchain.GetLastBlock().Index {
				if b.ComputeHash()[0:1] != "0" {
					fmt.Printf("H : %s\n", b.ComputeHash())
					b.Nonce++
				} else {
					b.Hash = b.ComputeHash()
					break
				}
			}
			if b.ComputeHash()[0:1] == "0" {
				break
			}
			txnNb = len(n.mempool)
		}

		log.Printf("Found good nonce for block!")
		n.blockchain.Lock()
		n.execute(b)
		b.StateRoot = n.blockchain.GetStateRoot()
		n.broadcastBlock(b)
		n.blockchain.Unlock()
		n.blockchain.AddBlock(b)
	}
}

func (n *Node) handle(conn net.Conn) {
	for {
		m := new(Message)
		err := json.NewDecoder(conn).Decode(m)
		if err != nil {
			log.Printf("Error decoding message from %s, closing connection", conn.RemoteAddr())
			log.Println(err)
			conn.Close()
			break
		}

		log.Printf("Received RPC : %s\nwith data : %s\nfrom : %s", m.Rpc, m.JSON, conn.RemoteAddr())
		log.Printf("Adding %s as a peer", conn.RemoteAddr()) //can know if this peer was already in peer list
		n.peers.LoadOrStore(conn, true)
		switch m.Rpc {
		case "blockchainrequest":
			n.blockchainRequestHandler(conn)
		case "blockchainreception":
			n.blockchainReceptionHandler(m.JSON)
		case "blockreception":
			n.blockReceptionHandler(m.JSON)
		case "transactionrequest":
			n.transactionRequestHandler(m.JSON)
		default:
			log.Printf("Remote procedure call %s does not exist on this client, ignoring...", m.Rpc)
		}
	}
}

func (n *Node) send(conn net.Conn, m *Message) {
	log.Printf("Sending RPC : %s\nwith data : %s\nto : %s", m.Rpc, m.JSON, conn.RemoteAddr())
	err := json.NewEncoder(conn).Encode(m)
	if err != nil {
		log.Printf("Error enconding message going to %s", conn.RemoteAddr())
		log.Println(err)
	}
}

func (n *Node) execute(b *blockchain.Block) {
	for _, t := range b.Txns {
		if acc, exists := n.blockchain.Accounts[t.Sender]; exists {
			acc.Balance -= t.Amount
		} else {
			acc = &blockchain.Account{
				Address: t.Receiver,
				Balance: -t.Amount,
			}
			n.blockchain.Accounts[t.Receiver] = acc
		}

		if acc, exists := n.blockchain.Accounts[t.Receiver]; exists {
			acc.Balance += t.Amount
		} else {
			acc = &blockchain.Account{
				Address: t.Receiver,
				Balance: t.Amount,
			}
			n.blockchain.Accounts[t.Receiver] = acc
		}
		//safe?
		delete(n.mempool, t.Hash[:])
	}

	if acc, exists := n.blockchain.Accounts[b.MinerAddress]; exists {
		acc.Balance += blockchain.BLOCK_REWARD
	} else {
		acc = &blockchain.Account{
			Address: b.MinerAddress,
			Balance: blockchain.BLOCK_REWARD,
		}
		n.blockchain.Accounts[b.MinerAddress] = acc
	}
}

func (n *Node) validateBlock(b *blockchain.Block) {
	if b.ComputeHash()[0:1] == "0" && b.Index == n.blockchain.GetLastBlock().Index+1 {
		log.Println("Received block validated, executing transactions...")
		n.execute(b)

		if n.blockchain.GetStateRoot() != b.StateRoot {
			log.Println("Printing state : ")
			n.printState()

			log.Printf("local SR after execution : \n%s\nBlock SR : \n%s", n.blockchain.GetStateRoot(), b.StateRoot)
			log.Println("State roots don't match, requesting bc...")

			n.peers.Range(func(key, value interface{}) bool {
				isActive := value.(bool)
				conn := key.(net.Conn)
				if isActive {
					n.requestBlockchain(conn)
					return false
				}
				return true
			})
		} else {
			n.blockchain.AddBlock(b)
		}
	} else {
		log.Println("Received block not valid, ignoring...")
	}
}

func (n *Node) validateBlockchain(bc *blockchain.Blockchain) {
	if bc.GetLastBlock().Index > n.blockchain.GetLastBlock().Index {
		log.Println("Received blockchain has higher index, validating blockchain...")
		if bc.IsValid() {
			log.Println("Received blockchain is longer and valid, replacing blockchain...")
			n.blockchain.ReplaceChain(bc)
		} else {
			log.Println("Received blockchain is invalid! ignoring...")
		}
	} else {
		log.Println("Received blockchain has lower or equal index, ignoring...")
	}
}

func (n *Node) getBlockchainAsMessage() (*Message, error) {
	n.blockchain.Lock()
	defer n.blockchain.Unlock()
	blockchainData, err := json.Marshal(n.blockchain)
	if err != nil {
		log.Println("Error while encoding blockchain")
		log.Println(err)
	} else {
		return &Message{
			Rpc:  "blockchainreception",
			JSON: blockchainData,
		}, nil
	}
	return nil, err
}

func (n *Node) requestBlockchain(conn net.Conn) {
	msg := &Message{
		Rpc:  "blockchainrequest",
		JSON: nil,
	}
	n.send(conn, msg)
}

func (n *Node) Init(target *string, keys *string) {
	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", n.listenPort))
	if err != nil {
		log.Println("Error starting listener")
		log.Println(err)
	}
	if *keys != "" {
		log.Println("Retrieving keys from file : ", *keys)
		data, err := os.ReadFile(*keys)
		if err != nil {
			log.Println("Error opening file")
			log.Println(err)
		}
		n.sigTree = crypto.UnmarshalJSON(data)
		if err != nil {
			log.Println("Error decoding data")
			log.Println(err)
		}
		log.Println("Retrieved MSS with public key : ", hex.EncodeToString(n.sigTree.GetPublicKey()))
	} else {
		log.Println("Generating new keys, storing to disk...")
		n.sigTree = crypto.NewMSS()
		name := hex.EncodeToString(n.sigTree.GetPublicKey()) + ".txt"
		keyData, err := n.sigTree.MarshalJSON()
		if err != nil {
			log.Println("Error encoding key data")
			log.Println(err)
		}
		err = os.WriteFile(name, keyData, 0644)
		if err != nil {
			log.Println("Error writing data to file")
			log.Println(err)
		}
	}

	n.account = &blockchain.Account{
		Address: hex.EncodeToString(n.sigTree.GetPublicKey()),
		Balance: 0,
	}
	n.blockchain.Init(n.account)
	n.mempool = make(map[string]blockchain.Transaction)
	if *target != "" {
		log.Printf("Trying to add peer %s", *target)
		conn, err := net.DialTimeout("tcp", *target, DIALTIMEOUT)
		if err != nil {
			log.Printf("Error dialing target %s", *target)
			log.Println(err)
		}
		n.peers.Store(conn, true)
		log.Printf("Added peer %s", *target)
		go n.handle(conn)
		n.requestBlockchain(conn)
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

	//TODO : remove this
	if *target != "" {
		go n.simulateLocalTxns()
		go n.simulateTxnRq()
	}
}

func (n *Node) simulateLocalTxns() {
	time.Sleep(time.Second)
	log.Println("Simulating local txns...")
	t := &blockchain.Transaction{
		Sender:    hex.EncodeToString(n.sigTree.GetPublicKey()),
		Receiver:  "13e1cf6b5b61055550fa30be5c64b4a3420208a34645d90093047495abfab583",
		Amount:    1,
		Timestamp: time.Now(),
	}
	t.Hash = t.ComputeHash()
	hash := make([]byte, 32)
	hex.Decode(hash, []byte(t.Hash))
	t.Signature = n.sigTree.Sign(*(*[32]byte)(hash))

	transactionData, err := json.Marshal(t)
	if err != nil {
		log.Println("Error encoding transaction data")
		log.Println(err)
	}
	n.transactionRequestHandler(transactionData)
}

//send txn request to send one coin to target, 5 times, to each peer
func (n *Node) simulateTxnRq() {
	time.Sleep(time.Second * 10)
	log.Println("Simulating txns")
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		n.peers.Range(func(k, v interface{}) bool {
			conn := k.(net.Conn)
			isValid := v.(bool)
			if isValid {
				t := &blockchain.Transaction{
					Sender:    hex.EncodeToString(n.sigTree.GetPublicKey()),
					Receiver:  "13e1cf6b5b61055550fa30be5c64b4a3420208a34645d90093047495abfab583",
					Amount:    1,
					Timestamp: time.Now(),
				}
				t.Hash = t.ComputeHash()
				hash := make([]byte, 32)
				hex.Decode(hash, []byte(t.Hash))
				t.Signature = n.sigTree.Sign(*(*[32]byte)(hash))

				transactionData, err := json.Marshal(t)
				if err != nil {
					log.Println("Error encoding transaction data")
					log.Println(err)
				}

				m := &Message{
					Rpc:  "transactionrequest",
					JSON: transactionData,
				}
				n.send(conn, m)
			}
			return true
		})
	}
}
