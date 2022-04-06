//TODO : save account to file

package main

import (
	"flag"
	"ketcoin/src/p2p"
	"log"
)

func main() {
	initNode()
}

func initNode() {
	listenPort := flag.Int("l", 0, "Port to listen on for new connections")
	target := flag.String("t", "", "Target peer to connect to at first")
	keys := flag.String("k", "", "File containing key information in JSON format")

	flag.Parse()

	if *listenPort == 0 {
		log.Fatal("Please provide a port to listen on with -l")
	}

	node := p2p.MakeNode(uint16(*listenPort))
	node.Init(target, keys)
	go node.Start()

	log.Printf("Try connecting to this node using \"./src -l %d -t 127.0.0.1:%d\"", *listenPort+1, *listenPort)
	select {}
}
