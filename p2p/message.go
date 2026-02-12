package p2p

import "net"

// RPC holds any arbitrary
// data that is being sent over each transport
// betweeen two nodes
type RPC struct {
	From    net.Addr
	Payload []byte
}
