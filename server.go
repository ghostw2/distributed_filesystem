package main

import (
	"bytes"
	"distributed-filesystem/p2p"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
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
			var p Payload
			if err := gob.NewDecoder(bytes.NewReader([]byte(msg.Payload))).Decode(&p); err != nil {
				fmt.Printf("error decoding the payload %v", err)
				continue
			}
			fmt.Printf("received a new payload with key %v and data %v \n", p.Key, string(p.Data))
			if err := f.handlePayload(&p); err != nil {
				fmt.Printf("error writing the payload to the store %v", err)
			}
		case <-f.quitChannel:
			return
		}
	}
}
func (f *FileServer) handlePayload(p *Payload) error {
	return f.store.Write(p.Key, bytes.NewReader(p.Data))
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
	f.BootstrapNetwork()
	go f.Loop()
	return nil
}
func (f *FileServer) Stop() {
	close(f.quitChannel)
	//f.quitChannel <- struct{}{}
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
	tee := io.TeeReader(r, buff)

	// Store the date to disk
	if err := f.store.Write(key, tee); err != nil {
		return err
	}

	payload := &Payload{
		Key:  key,
		Data: buff.Bytes(),
	}
	// Broadcast to. all other known peers that you have the date

	return f.Broadcast(payload)
}
