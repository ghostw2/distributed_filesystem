package p2p

// HandshakeFunc is ... ?
type HandshakerFunc func(Peer) error

func NopeHandshakeFunc(Peer) error { return nil }
