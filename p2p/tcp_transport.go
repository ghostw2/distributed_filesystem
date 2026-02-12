package p2p

import (
	"fmt"
	"net"
)

// TCPPeer represents the remote node over a TCP established connection
type TCPPeer struct {
	//coon is the underlaying connection of the peer
	conn net.Conn
	//if we dail a conn -> outbond
	//if we accept a conn and retrive -> outband false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbond bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbond,
	}
}

// Close implements the Peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
}
type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakerFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

// Consume implements the transport interface
// which will return  a read-only channnel
// for reading the other messages recived form another peer
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) ListenAndAccept() error {
	var (
		err error
	)
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()
	return nil

}
func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP Accept error: %s\n", err)
		}
		fmt.Printf("New incomming connection %v \n", conn)
		go t.handleConn(conn)
		fmt.Println("Affer the first handle conn")
	}

}

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error

	defer func() {
		fmt.Printf("dropping peer connection : %v \n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)
	if err = t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("TCP error handshake : %s \n", err)
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			fmt.Printf("this is the errr %v \n", err)
			return
		}
	}
	// Read Loop
	rpc := RPC{}
	for {
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Printf("this is en decoding error it should stop %v", err)
			break
		}
		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}

}
