package p2p

import "errors"

// ErrInvalidHanshake is returned if the handshake
// between local and remote  node is not established
var ErrInvalidHandshake = errors.New("invalid handshake")

// HandshakeFunc is ... ?
type HandshakerFunc func(Peer) error

func NopeHandshakeFunc(Peer) error { return nil }
