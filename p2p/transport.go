package p2p

// Peer is an interface that represents the remote node
type Peer interface {
	Close() error
}

// Transport is everything that handles communication
// between nodes in your network.(This can be of the form
// TCP,UDP,Websockets)
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
}
