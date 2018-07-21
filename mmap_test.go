package mmap

import (
	"bytes"
	"io/ioutil"
	"os"
	"syscall"
	"testing"
)

var (
	testData = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	testPath = "/tmp/m.txt"
)

func init() {
	f, err := os.OpenFile(testPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.Write(testData)
}

func TestUnmap(t *testing.T) {
	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		t.Errorf("error in opening file :: %v", err)
	}
	defer f.Close()

	m, err := NewSharedFileMmap(f, 0, len(testData), syscall.SYS_READ)
	if err != nil {
		t.Errorf("error in mapping :: %v", err)
	}

	if err := m.Unmap(); err != nil {
		t.Errorf("error in unmapping :: %v", err)
	}
}

func TestReadWrite(t *testing.T) {
	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	m, err := NewSharedFileMmap(f, 0, len(testData), syscall.SYS_READ)
	if err != nil {
		t.Errorf("error in mapping :: %v", err)
	}
	defer m.Unmap()

	// Read
	data := make([]byte, len(testData))
	if _, err := m.Read(data, 0); err != nil {
		t.Errorf("error in reading :: %v", err)
	}
	if !bytes.Equal(testData, data) {
		t.Errorf("mapped data is not equal testData: %v, %v", data, testData)
	}

	// Read slice bigger than mapped region after offset
	lenData := len(testData) + 10
	data = make([]byte, lenData)
	if _, err := m.Read(data, 0); err != nil {
		t.Errorf("error in reading :: %v", err)
	}
	if !bytes.Equal(testData, data[:len(testData)]) {
		t.Errorf("mapped data is not equal testData: %v, %v", data[:len(testData)], testData)
	}
	if !bytes.Equal(data[len(testData):], make([]byte, 10)) {
		t.Errorf("mapped data is not equal testData: %v, %v", data[:len(testData)], testData)
	}

	// Read offset larger than size of mapped region
	if _, err := m.Read(data, 100); err != ErrIndexOutOfBound {
		t.Errorf("unexpected error in reading from mmaped region :: %v", err)
	}

	// Write
	if _, err := m.Write([]byte("a"), 9); err != nil {
		t.Error(err)
	}
	m.Flush(syscall.SYS_SYNC)
	f1, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	fileData, err := ioutil.ReadAll(f1)
	if err != nil {
		t.Errorf("error in reading file: %s", err)
	}
	f1.Close()
	if !bytes.Equal(fileData, []byte("012345678aABCDEFGHIJKLMNOPQRSTUVWXYZ")) {
		t.Errorf("no modification in file: %v", string(fileData))
	}

	// Write slice bigger than mapped region after offset
	if _, err := m.Write([]byte("abc"), 34); err != nil {
		t.Error(err)
	}
	m.Flush(syscall.SYS_SYNC)
	f2, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	fileData, err = ioutil.ReadAll(f2)
	if err != nil {
		t.Errorf("error in reading file: %s", err)
	}
	f2.Close()
	if !bytes.Equal(fileData, []byte("012345678aABCDEFGHIJKLMNOPQRSTUVWXab")) {
		t.Errorf("no modification in file: %v", string(fileData))
	}

	// Write offset larger than size of mapped region
	if _, err := m.Write([]byte("a"), 100); err != ErrIndexOutOfBound {
		t.Errorf("unexpected error in writing to mmaped region :: %v", err)
	}
}

func TestAdvise(t *testing.T) {
	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	m, err := NewSharedFileMmap(f, 0, len(testData), syscall.SYS_READ)
	if err != nil {
		t.Errorf("error in mapping :: %v", err)
	}
	defer m.Unmap()

	m.Advise(syscall.MADV_SEQUENTIAL)
}

func TestLockUnlock(t *testing.T) {
	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	m, err := NewSharedFileMmap(f, 0, len(testData), syscall.SYS_READ)
	if err != nil {
		t.Errorf("error in mapping :: %v", err)
	}
	defer m.Unmap()

	m.Lock()
	m.Unlock()
}
