package main

import (
	"bytes"
	"distributed-filesystem/p2p"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
	"time"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	TCPTransportOpts  p2p.TCPTransportOpts
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts
	Peers    map[string]p2p.Peer
	PeerLock sync.Mutex
	store    *Store

	quitChannel chan struct{}
}

func (f *FileServer) Loop() {
	defer func() {
		if err := f.Transport.Close(); err != nil {
			fmt.Printf("failed to stop the server error:%v", err)
		}
		fmt.Println("file server stopped")
	}()
	for {
		select {
		case msg := <-f.Transport.Consume():

			var p Message
			if err := gob.NewDecoder(bytes.NewReader([]byte(msg.Payload))).Decode(&p); err != nil {
				fmt.Printf("error decoding the payload %v", err)
				continue
			}
			fmt.Printf("the message is coming from %v", msg.From.String())

			if err := f.handleMessage(msg.From.String(), &p); err != nil {
				fmt.Printf("error handling message %v", err)
			}
			peer := f.Peers[msg.From.String()]
			if peer == nil {
				panic("peer not found for the incoming message")
			}
			b := make([]byte, 1024)
			n, err := peer.Read(b)
			if err != nil {
				fmt.Printf("error reading from the peer %v error: %v \n", peer.RemoteAddr(), err)
			}
			fmt.Printf("the message read from the peer is %s \n", string(b[:n]))

			// if err := f.handlePayload(&p); err != nil {
			// 	fmt.Printf("error writing the payload to the store %v", err)
			// }
			peer.Done()

		case <-f.quitChannel:
			return
		}
	}
}
func (f *FileServer) handlePayload(p *Payload) error {
	return f.store.Write(p.Key, bytes.NewReader(p.Data))
}
func (f *FileServer) handleMessage(from string, m *Message) error {
	switch payload := m.Payload.(type) {
	case *MessageStoreFile:

		return f.handleMessageStoreFile(from, payload)
	default:
		return fmt.Errorf("unknown message type")
	}
}
func (f *FileServer) handleMessageStoreFile(from string, m *MessageStoreFile) error {
	peer := f.Peers[from]
	if peer == nil {
		panic("peer not found for the incoming message")
	}

	fmt.Printf("handling the message store file with contents :% with key %v \n", from, m.Key)
	if err := f.store.Write(m.Key, io.LimitReader(peer, 10)); err != nil {
		return err
	}
	peer.Done()
	return nil
}
func NewFileServer(opts FileServerOpts) *FileServer {
	return &FileServer{
		FileServerOpts: opts,
		quitChannel:    make(chan struct{}),
		store: NewStore(&StoreOpts{
			Root:          opts.StorageRoot,
			TransformFunc: opts.PathTransformFunc,
		}),
		Peers: make(map[string]p2p.Peer),
	}
}

func (f *FileServer) Start() error {
	if err := f.Transport.ListenAndAccept(); err != nil {
		return err
	}
	gob.Register(&MessageStoreFile{})
	f.BootstrapNetwork()
	go f.Loop()
	return nil
}
func (f *FileServer) Stop() {
	close(f.quitChannel)
}
func (f *FileServer) Store(key string, r io.Reader) error {

	return f.store.Write(key, r)
}
func (f *FileServer) BootstrapNetwork() error {
	for _, node := range f.BootstrapNodes {
		if len(node) == 0 {
			continue
		}
		go func(node string) {
			if err := f.Transport.Dial(node); err != nil {
				fmt.Printf("error dialing : %v \n", err)
			}
		}(node)
	}
	return nil
}
func (f *FileServer) OnPeer(p p2p.Peer) error {
	f.PeerLock.Lock()
	defer f.PeerLock.Unlock()
	f.Peers[p.RemoteAddr().String()] = p
	fmt.Printf("added new peer %v \n", p.RemoteAddr())
	return nil
}

type Payload struct {
	Key  string
	Data []byte
}
type Message struct {
	Payload any
}
type MessageStoreFile struct {
	Key string
}

func (f *FileServer) Broadcast(p *Payload) error {
	peersSlice := []io.Writer{}

	for _, peer := range f.Peers {
		peersSlice = append(peersSlice, peer)
	}
	mw := io.MultiWriter(peersSlice...)
	//fmt.Printf("broadcasting to %v peers the payload with payload %v \n", len(peersSlice), p)
	return gob.NewEncoder(mw).Encode(p)
}

func (f *FileServer) StoreDate(key string, r io.Reader) error {
	// Read the data from the reader and store it in a buffer
	buff := new(bytes.Buffer)
	// tee := io.TeeReader(r, buff)

	// Store the date to disk
	// if err := f.store.Write(key, tee); err != nil {
	// 	return err
	// }

	payload := &Message{
		Payload: &MessageStoreFile{
			Key: key,
		},
	}
	// Broadcast to. all other known peers that you have the date
	if err := gob.NewEncoder(buff).Encode(payload); err != nil {
		fmt.Printf("theres is an error here mate %v\n", err)
		return err
	}
	for _, peer := range f.Peers {
		if err := peer.Send(buff.Bytes()); err != nil {
			fmt.Printf("error broadcasting to peer %v error: %v \n", peer.RemoteAddr(), err)
		}
	}
	time.Sleep(time.Second * 3)

	p := []byte("THIS IS A BIG FILE")

	for _, peer := range f.Peers {
		if err := peer.Send(p); err != nil {
			fmt.Printf("error broadcasting to peer %v error: %v \n", peer.RemoteAddr(), err)
		}
	}
	// fmt.Printf("%s\n", string(p))

	return nil
	// return f.Broadcast(payload)
}
