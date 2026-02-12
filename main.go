package main

import (
	"distributed-filesystem/p2p"
	"fmt"
	"log"
)

func OnPeer(p2p.Peer) error {
	fmt.Printf("doing some logic with the peer outside of the TCPTransport")
	// return fmt.Errorf("failed the on onPeer func")
	return nil
}

func main() {
	opts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NopeHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransport(opts)
	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("message %v \n", msg)
		}
	}()
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatalf("Error in listening and accepting :%v\n", err)
	}
	fmt.Println("we are a go")

	select {}
}
