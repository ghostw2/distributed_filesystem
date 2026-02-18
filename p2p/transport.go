package p2p

import "net"

// Peer is an interface that represents the remote node
type Peer interface {
	net.Conn
	Send([]byte) error
}

// Transport is everything that handles communication
// between nodes in your network.(This can be of the form
// TCP,UDP,Websockets)
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
	Addr() string
}
