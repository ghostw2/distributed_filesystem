package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

func CASPathTransformFunc(key string) *PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}
	return &PathKey{
		Pathname: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

type Store struct {
	StoreOpts
}
type StoreOpts struct {
	TransformFunc PathTransfromFunc
}
type PathTransfromFunc func(string) *PathKey

var DefaultPathTransformFunc = func(key string) *PathKey {
	return &PathKey{
		Pathname: key,
		Filename: key,
	}
}

type PathKey struct {
	Pathname string
	Filename string
}

func (p *PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.Pathname, p.Filename)
}
func NewStorage(opts *StoreOpts) *Store {
	return &Store{
		StoreOpts: *opts,
	}
}
func (s *Store) Read(key string) io.Reader {
	f, err := s.readStream(key)
	if err != nil {
		return nil
	}
	defer f.Close()

	buff := new(bytes.Buffer)

	_, err = io.Copy(buff, f)
	if err != nil {
		return nil
	}
	return buff
}
func (s *Store) HasFile(key string) bool {
	pathKey := s.TransformFunc(key)
	_, err := os.Stat(pathKey.FullPath())
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return true

}
func (s *Store) Delete(key string) error {
	pathKey := s.TransformFunc(key)
	fmt.Printf("\n removed :%s ", pathKey.FullPath())
	return os.RemoveAll(pathKey.FullPath())
}
func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.TransformFunc(key)
	// f, err := os.ReadFile(pathKey.FullPath())
	return os.Open(pathKey.FullPath())
}
func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.TransformFunc(key)
	err := os.MkdirAll(pathKey.Pathname, os.ModePerm)
	if err != nil {
		return err
	}

	fullPathName := pathKey.FullPath()
	f, err := os.Create(fullPathName)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	log.Printf("written (%d) bytes to disk :%s", n, fullPathName)
	return nil
}
