package main

import (
	"distributed-filesystem/p2p"
	"fmt"
	"log"
)

func main() {
	tr := p2p.NewTCPTransport(":3000")
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatalf("Error in listening and accepting :%v\n", err)
	}
	fmt.Println("we are a go")

	select {}
}
