package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func newEncryptionKey() []byte {
	keyBuff := make([]byte, 32)
	io.ReadFull(rand.Reader, keyBuff)
	return keyBuff
}
func copyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	// Read the iV from the given io.Reader which, in our case
	// should be the block.BlockSize() we read.
	iv := make([]byte, block.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return 0, nil
	}
	var (
		buff   = make([]byte, 32*1024)
		stream = cipher.NewCTR(block, iv)
	)
	for {
		n, err := src.Read(buff)
		if n > 0 {
			stream.XORKeyStream(buff, buff[:n])
			if _, err := dst.Write(buff[:n]); err != nil {
				return 0, nil
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return 0, nil
}
func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	iv := make([]byte, block.BlockSize()) //16 bytes
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	//Prepend the IV to the file
	if _, err := dst.Write(iv); err != nil {
		return 0, nil
	}

	var (
		buff   = make([]byte, 32*1024)
		stream = cipher.NewCTR(block, iv)
	)
	for {
		n, err := src.Read(buff)
		if n > 0 {
			stream.XORKeyStream(buff, buff[:n])
			if _, err := dst.Write(buff[:n]); err != nil {
				return 0, nil
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return 0, nil
}
