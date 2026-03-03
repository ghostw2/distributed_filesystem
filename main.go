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
	// var err error
	// var r io.Reader
	// r, err = fs_2.Get("_secure_token")
	// if err != nil {
	// 	fmt.Printf("error getting the file %v", err)
	// }
	// if r != nil {
	// 	var b []byte
	// 	b, err = io.ReadAll(r)
	// 	if err != nil {
	// 		fmt.Printf("error reading the file content %v", err)
	// 	}
	// 	fmt.Printf("the content of the file is %s \n", string(b))
	// } else {
	// 	fmt.Printf("the file is not found in the network \n")
	// }
	for i := 0; i < 15; i++ {
		data := bytes.NewReader([]byte("death to ming" + fmt.Sprintf("%d", i)))
		fmt.Printf("Error value %v", fs_2.Store("_secure_token_"+fmt.Sprintf("%d", i), data))

		time.Sleep(1 * time.Millisecond)
	}
	select {}

}
