package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, v any) error {
	return gob.NewDecoder(r).Decode(v)
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	peekBuff := make([]byte, 1)
	_, err := r.Read(peekBuff)
	if err != nil {
		return err
	}
	stream := peekBuff[0] == IncomingStream
	// In case of a stream we do not decode what is being sent over the wire,
	// instead we just set the stream flag to true
	if stream {
		msg.Stream = true
		return nil
	}

	buf := make([]byte, 1028)

	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	msg.Payload = buf[:n]
	return nil
}
