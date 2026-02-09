package p2p

import (
	"fmt"
	"net"
	"sync"
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

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	shakeHands    HandshakerFunc
	decoder       Decoder
	mu            sync.RWMutex
	peers         map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		shakeHands:    NopeHandshakeFunc,
		listenAddress: listenAddr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var (
		err error
	)
	t.listener, err = net.Listen("tcp", t.listenAddress)
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
	}

}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	if err := t.shakeHands(peer); err != nil {

	}
	lenDecodeError := 0
	// Read Loop
	msg := &Temp{}
	for {
		if err := t.decoder.Decode(conn, msg); err != nil {
			lenDecodeError++
			if lenDecodeError > 5 {
				fmt.Printf("TCP error once again : %s \n", err)
				continue
			}
		}
	}
}
