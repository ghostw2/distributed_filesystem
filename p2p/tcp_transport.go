package p2p

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

// TCPPeer represents the remote node over a TCP established connection
type TCPPeer struct {
	//coon is the underlying connection of the peer
	net.Conn
	//if we dial a conn -> outbound
	//if we accept a conn and retrieve -> outbound false
	outbound bool

	wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Write(b)
	return err
}
func (p *TCPPeer) Done() {
	p.wg.Done()
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
// which will return  a read-only channel
// for reading the other messages received form another peer
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}
func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

func (t *TCPTransport) ListenAndAccept() error {
	var (
		err error
	)
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	fmt.Printf("Listening on port :%v \n", t.ListenAddr)
	go t.startAcceptLoop()
	return nil

}
func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP Accept error: %s\n", err)
		}
		if errors.Is(err, net.ErrClosed) {
			return
		}
		go t.handleConn(conn, false)
	}

}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		fmt.Printf("dropping peer connection : %v \n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)
	if err = t.HandshakeFunc(peer); err != nil {
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
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Printf("this is en decoding error it should stop %v", err)
			break
		}
		rpc.From = conn.RemoteAddr()
		peer.wg.Add(1)
		fmt.Printf("Sending data form peer %v \n", peer)
		t.rpcch <- rpc

		peer.wg.Wait()
		fmt.Println("Done with the stream")
	}

}

// Close implements the Close method of the transport interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial implements the transport interface
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	fmt.Printf("dialed new peer %v \n", addr)
	return nil
}
