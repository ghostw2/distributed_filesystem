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
	"path/filepath"
	"strings"
)

const defaultRoot = "storage"

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
	Root          string
	TransformFunc PathTransformFunc
}
type PathTransformFunc func(string) *PathKey

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

func (p *PathKey) FullPath(root string) string {
	return fmt.Sprintf("%s/%s/%s", root, p.Pathname, p.Filename)
}
func NewStorage(opts *StoreOpts) *Store {
	if opts.TransformFunc == nil {
		opts.TransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultRoot
	}
	return &Store{
		StoreOpts: *opts,
	}
}
func (s *Store) FullPathName(key string) []string {
	pathKey := s.TransformFunc(key)
	paths := strings.Split(pathKey.Pathname, "/")
	return paths
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
	_, err := os.Stat(pathKey.FullPath(s.Root))
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return true

}
func (s *Store) Delete(key string) error {
	pathKey := s.TransformFunc(key)
	fmt.Printf("\n removed :%s ", pathKey.FullPath(s.Root))
	err := os.RemoveAll(pathKey.FullPath(s.Root))
	if err != nil {
		return err
	}
	// Clean up empty parent directories
	paths := strings.Split(pathKey.Pathname, "/")
	for i := len(paths); i > 0; i-- {

		// Build the current directory path
		currentPath := strings.Join(paths[:i], "/")

		err := os.Remove(currentPath)
		if err != nil {
			fmt.Printf("error in  Remove :%v", err)
			break
		}
		fmt.Printf("Removed empty directory: %s\n", currentPath)
	}
	return nil
}
func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.TransformFunc(key)
	// f, err := os.ReadFile(pathKey.FullPath())
	return os.Open(pathKey.FullPath(s.Root))
}
func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.TransformFunc(key)
	err := os.MkdirAll(filepath.Join(s.Root, pathKey.Pathname), os.ModePerm)
	if err != nil {
		return err
	}

	fullPathName := pathKey.FullPath(s.Root)
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
