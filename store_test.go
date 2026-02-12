package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransform(t *testing.T) {
	key := "funnypicofmenri"
	pathname := CASPathTransformFunc(key)
	expectedFilename := "524cf9bf60f19f3583d2ee55842886494de6f6ee"
	expectedPath := "524cf/9bf60/f19f3/583d2/ee558/42886/494de/6f6ee"
	if expectedPath != pathname.Pathname {
		t.Errorf("the value of the pathname does not match the expected\n Expcted :%v \n Actual:%v \n", expectedPath, pathname.Pathname)
	}
	if expectedFilename != pathname.Filename {
		fmt.Println(pathname.Filename)
		t.Errorf("the value of the Filename does not match the expected\n Expcted :%v \n Actual:%v \n", expectedPath, pathname.Filename)
	}
}
func TestStoreDelete(t *testing.T) {
	opts := &StoreOpts{
		TransformFunc: CASPathTransformFunc,
	}
	store := NewStorage(opts)
	key := "my_super_sesscure_dkey"
	Val := []byte("somsse jpg bytes")
	data := bytes.NewReader(Val)
	if err := store.writeStream(key, data); err != nil {
		t.Error(err)
	}
	if !store.HasFile(key) {
		t.Error("file does not exits")
	}
	store.Delete(key)
	if store.HasFile(key) {
		t.Error("file still exits")
	}

}
func TestStorage(t *testing.T) {
	opts := &StoreOpts{
		TransformFunc: CASPathTransformFunc,
	}
	storage := NewStorage(opts)
	key := "my_super_secure_dkey"
	Val := []byte("some jpg bytes")
	data := bytes.NewReader(Val)
	if err := storage.writeStream(key, data); err != nil {
		t.Error(err)
	}

	fileContent := storage.Read(key)

	f, err := io.ReadAll(fileContent)

	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(f, Val) {
		t.Errorf("\nthe expected value :%v \n the actual value :%v",
			Val, f)
	}
	storage.Delete(key)
	fileContent = storage.Read(key)
	if fileContent != nil {
		t.Errorf("\n the file contents are %v", fileContent)
	}

}
