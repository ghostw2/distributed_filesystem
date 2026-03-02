package main

import (
	"bytes"
	"distributed-filesystem/p2p"
	"encoding/gob"
	"fmt"
	"io"
	"log"
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
			fmt.Printf("received a message from peer %v with payload %v \n", msg.From.String(), string(msg.Payload))
			var p Message
			if err := gob.NewDecoder(bytes.NewReader([]byte(msg.Payload))).Decode(&p); err != nil {
				fmt.Printf("error decoding the payload %v", err)
				continue
			}
			fmt.Printf("the message is coming from %v and the value is %v \n", msg.From.String(), &p)

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
		case <-f.quitChannel:
			return
		}
	}
}

func (f *FileServer) handleMessage(from string, m *Message) error {
	switch payload := m.Payload.(type) {
	case *MessageStoreFile:
		return f.handleMessageStoreFile(from, payload)
	case *MessageGetFile:
		return f.handleMessageGetFile(from, payload)
	default:
		return fmt.Errorf("unknown message type")
	}
}
func (f *FileServer) handleMessageGetFile(from string, m *MessageGetFile) error {
	fmt.Printf("Handling request for a file form peer %v", from)
	peer := f.Peers[from]
	if peer == nil {
		return fmt.Errorf("peer not found")
	}
	data, err := f.Get(m.Key)
	if err != nil {
		return err
	}
	fileBuff := new(bytes.Buffer)
	io.Copy(fileBuff, data)
	msg := &Message{
		Payload: &MessageStoreFile{
			Key:  m.Key,
			Size: int64(fileBuff.Len()),
		},
	}
	headerBuff := new(bytes.Buffer)
	if err := gob.NewEncoder(headerBuff).Encode(msg); err != nil {
		return err
	}
	if err := peer.Send(headerBuff.Bytes()); err != nil {
		return err
	}
	if err := peer.Send(fileBuff.Bytes()); err != nil {
		return err
	}
	return nil
}
func (f *FileServer) handleMessageStoreFile(from string, m *MessageStoreFile) error {
	peer := f.Peers[from]
	if peer == nil {
		panic("peer not found for the incoming message")
	}

	n, err := f.store.Write(m.Key, io.LimitReader(peer, m.Size))
	if err != nil {
		return err
	}
	log.Printf("finished writing the file with key %s and size %d bytes to disk", m.Key, n)

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
	gob.Register(&MessageGetFile{})
	f.BootstrapNetwork()
	go f.Loop()
	return nil
}
func (f *FileServer) Stop() {
	close(f.quitChannel)
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

type Message struct {
	Payload any
}
type MessageStoreFile struct {
	Key  string
	Size int64
}
type MessageGetFile struct {
	Key string
}

func (f *FileServer) Stream(m *Message) error {
	peersSlice := []io.Writer{}
	for _, peer := range f.Peers {
		peersSlice = append(peersSlice, peer)
	}
	mw := io.MultiWriter(peersSlice...)
	return gob.NewEncoder(mw).Encode(m)
}
func (f *FileServer) Broadcast(m *Message) error {
	fmt.Printf("Broadcasting  tha data %v", m.Payload)
	buff := new(bytes.Buffer)
	if err := gob.NewEncoder(buff).Encode(m); err != nil {
		return err
	}
	for _, peer := range f.Peers {
		if err := peer.Send(buff.Bytes()); err != nil {
			return err
		}
	}
	return nil
}
func (f *FileServer) Get(key string) (io.Reader, error) {
	if !f.store.HasFile(key) {
		fmt.Printf("we do not have the file with key %s in our local store, broadcasting to other peers \n", key)
		getMsg := &Message{
			Payload: &MessageGetFile{
				Key: key,
			},
		}
		err := f.Broadcast(getMsg)
		if err != nil {
			fmt.Printf("error broadcasting the get message to other peers %v\n", err)
			return nil, err
		}

		return nil, fmt.Errorf("file not found in network")

	}
	data := f.store.Read(key)
	if data == nil {
		return nil, fmt.Errorf("file not found")
	}
	return data, nil
}

func (f *FileServer) Store(key string, r io.Reader) error {
	// Read the data from the reader and store it in a buffer
	fileBuff := new(bytes.Buffer)
	io.Copy(fileBuff, r)
	// Store the date to disk
	var n int64
	n, err := f.store.Write(key, bytes.NewReader(fileBuff.Bytes()))
	if err != nil {
		return err
	}
	payload := &Message{
		Payload: &MessageStoreFile{
			Key:  key,
			Size: n,
		},
	}
	// Broadcast to. all other known peers that you have the date
	if err := f.Broadcast(payload); err != nil {
		fmt.Printf("theres is an error here mate %v\n", err)
		return err
	}
	time.Sleep(time.Second * 2)
	// TODO : we can use a MultiWriter here
	for _, peer := range f.Peers {
		if _, err := io.Copy(peer, bytes.NewReader(fileBuff.Bytes())); err != nil {
			fmt.Printf("error broadcasting to peer %v error: %v \n", peer.RemoteAddr(), err)
		}
	}
	return nil
}
