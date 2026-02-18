package main

import (
	"bytes"
	"distributed-filesystem/p2p"
	"fmt"
	"time"
)

func OnPeer(p2p.Peer) error {
	//fmt.Printf("doing some logic with the peer outside of the TCPTransport")
	// return fmt.Errorf("failed the on onPeer func")
	return nil
}
func makeFileServer(listenAddr string, bootstrapNodes []string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NopeHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fsOpts := FileServerOpts{
		StorageRoot:       listenAddr + "_storage",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    bootstrapNodes,
	}
	f := NewFileServer(fsOpts)
	tcpTransport.OnPeer = f.OnPeer
	return f
}
func main() {
	fs_1 := makeFileServer(":3000", nil)
	fs_2 := makeFileServer(":4000", []string{":3000"})

	go fs_1.Start()
	time.Sleep(1 * time.Second)
	go fs_2.Start()
	time.Sleep(2 * time.Second)
	data := bytes.NewReader([]byte("hello darkness my old friend"))
	fmt.Printf("Error value %v", fs_2.StoreDate("my_super_secure_token", data))
	select {}

}
