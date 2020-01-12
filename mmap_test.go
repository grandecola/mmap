package mmap

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"testing"
)

var (
	protPage = syscall.PROT_READ | syscall.PROT_WRITE
	testData = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	testPath = "/tmp/m.txt"
)

func setup(t *testing.T) {
	f, err := os.OpenFile(testPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("error in opening file :: %v", err)
	}

	if _, err := f.Write(testData); err != nil {
		t.Fatalf("error in writing to file :: %v", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("error in closing file :: %v", err)
	}
}

func tearDown(t *testing.T) {
	if err := os.Remove(testPath); err != nil {
		t.Fatalf("error in deleting file :: %v", err)
	}
}

func TestUnmap(t *testing.T) {
	setup(t)
	defer tearDown(t)

	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("error in opening file :: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("error in closing file :: %v", err)
		}
	}()

	m, err := NewSharedFileMmap(f, 0, len(testData), protPage)
	if err != nil {
		t.Fatalf("error in mapping :: %v", err)
	}

	if err := m.Unmap(); err != nil {
		t.Fatalf("error in unmapping :: %v", err)
	}
}

func TestReadWrite(t *testing.T) {
	setup(t)
	defer tearDown(t)

	f, errFile := os.OpenFile(testPath, os.O_RDWR, 0644)
	if errFile != nil {
		t.Fatalf("error in opening file :: %v", errFile)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("error in closing file :: %v", err)
		}
	}()

	m, errMmap := NewSharedFileMmap(f, 0, len(testData), protPage)
	if errMmap != nil {
		t.Fatalf("error in mapping :: %v", errMmap)
	}
	defer func() {
		if err := m.Unmap(); err != nil {
			t.Fatalf("error in calling unmap :: %v", err)
		}
	}()

	// Read
	data := make([]byte, len(testData))
	if _, err := m.ReadAt(data, 0); err != nil {
		t.Fatalf("error in reading :: %v", err)
	}
	if !bytes.Equal(testData, data) {
		t.Fatalf("mapped data is not equal testData: %v, %v", data, testData)
	}

	// Read slice bigger than mapped region after offset
	lenData := len(testData) + 10
	data = make([]byte, lenData)
	if _, err := m.ReadAt(data, 0); err != nil {
		t.Fatalf("error in reading :: %v", err)
	}
	if !bytes.Equal(testData, data[:len(testData)]) {
		t.Fatalf("mapped data is not equal testData: %v, %v", data[:len(testData)], testData)
	}
	if !bytes.Equal(data[len(testData):], make([]byte, 10)) {
		t.Fatalf("mapped data is not equal testData: %v, %v", data[:len(testData)], testData)
	}

	// Read offset larger than size of mapped region
	func() {
		defer func() {
			if err := recover(); err != ErrIndexOutOfBound {
				t.Fatalf("unexpected error in reading from mmaped region :: %v", err)
			}
		}()

		_, _ = m.ReadAt(data, 100)
	}()

	// Write
	if _, err := m.WriteAt([]byte("a"), 9); err != nil {
		t.Fatalf("error in writing to mapped area :: %v", err)
	}
	if err := m.Flush(syscall.MS_SYNC); err != nil {
		t.Fatalf("error in calling flush :: %v", err)
	}
	f1, errFile := os.OpenFile(testPath, os.O_RDWR, 0644)
	if errFile != nil {
		t.Fatalf("error in opening file :: %v", errFile)
	}
	fileData, errFile := ioutil.ReadAll(f1)
	if errFile != nil {
		t.Fatalf("error in reading file :: %s", errFile)
	}
	if err := f1.Close(); err != nil {
		t.Fatalf("error in closing file :: %v", err)
	}
	if !bytes.Equal(fileData, []byte("012345678aABCDEFGHIJKLMNOPQRSTUVWXYZ")) {
		t.Fatalf("no modification in file :: %v", string(fileData))
	}

	// Write slice bigger than mapped region after offset
	if _, err := m.WriteAt([]byte("abc"), 34); err != nil {
		t.Fatalf("error in writing to mapped region :: %v", err)
	}
	if err := m.Flush(syscall.MS_SYNC); err != nil {
		t.Fatalf("error in flushing mapped region :: %v", err)
	}
	f2, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("error in opening file :: %v", err)
	}
	fileData, err = ioutil.ReadAll(f2)
	if err != nil {
		t.Fatalf("error in reading file :: %s", err)
	}
	if err := f2.Close(); err != nil {
		t.Fatalf("error in closing file :: %s", err)
	}
	if !bytes.Equal(fileData, []byte("012345678aABCDEFGHIJKLMNOPQRSTUVWXab")) {
		t.Fatalf("no modification in file :: %v", string(fileData))
	}

	// Write offset larger than size of mapped region
	func() {
		defer func() {
			if err := recover(); err != ErrIndexOutOfBound {
				t.Fatalf("unexpected error in writing to mmaped region :: %v", err)
			}
		}()

		_, _ = m.WriteAt([]byte("a"), 100)
	}()
}

func TestReadStringAt(t *testing.T) {
	setup(t)
	defer tearDown(t)

	f, errFile := os.OpenFile(testPath, os.O_RDWR, 0644)
	if errFile != nil {
		t.Fatalf("error in opening file :: %v", errFile)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("error in closing file :: %v", err)
		}
	}()

	m, errMmap := NewSharedFileMmap(f, 0, len(testData), protPage)
	if errMmap != nil {
		t.Fatalf("error in mapping :: %v", errMmap)
	}
	defer func() {
		if err := m.Unmap(); err != nil {
			t.Fatalf("error in calling unmap :: %v", err)
		}
	}()

	// Read
	sb := &strings.Builder{}
	sb.Grow(len(testData))
	if n := m.ReadStringAt(sb, 0); n != len(testData) {
		t.Fatalf("expected to read string of length %v, read of length %v", len(testData), n)
	}
	if string(testData) != sb.String() {
		t.Fatalf("mapped data is not equal testData: %v, %v", sb.String(), string(testData))
	}

	// Read data smaller than the mapped region
	sb.Reset()
	sb.Grow(len(testData) - 2)
	if n := m.ReadStringAt(sb, 0); n != len(testData)-2 {
		t.Fatalf("expected to read string of length %v, read of length %v", len(testData), n)
	}
	expectedData := string(testData[:len(testData)-2])
	if expectedData != sb.String() {
		t.Fatalf("mapped data is not equal testData: %v, %v", sb.String(), expectedData)
	}

	// Read slice bigger than mapped region after offset
	sb.Reset()
	sb.Grow(len(testData) + 10)
	if n := m.ReadStringAt(sb, 0); n != len(testData) {
		t.Fatalf("expected to read string of length %v, read of length %v", len(testData), n)
	}
	if string(testData) != sb.String() {
		t.Fatalf("mapped data is not equal testData: %v, %v", sb.String(), string(testData))
	}

	// Read offset larger than size of mapped region
	func() {
		defer func() {
			if err := recover(); err != ErrIndexOutOfBound {
				t.Fatalf("unexpected error in reading from mmaped region :: %v", err)
			}
		}()

		sb.Reset()
		sb.Grow(10)
		_ = m.ReadStringAt(sb, 100)
	}()
}

func TestWriteStringAt(t *testing.T) {
	setup(t)
	defer tearDown(t)

	f, errFile := os.OpenFile(testPath, os.O_RDWR, 0644)
	if errFile != nil {
		t.Fatalf("error in opening file :: %v", errFile)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("error in closing file :: %v", err)
		}
	}()

	m, errMmap := NewSharedFileMmap(f, 0, len(testData), protPage)
	if errMmap != nil {
		t.Fatalf("error in mapping :: %v", errMmap)
	}
	defer func() {
		if err := m.Unmap(); err != nil {
			t.Fatalf("error in calling unmap :: %v", err)
		}
	}()

	// Write
	_ = m.WriteStringAt("a", 9)
	if err := m.Flush(syscall.MS_SYNC); err != nil {
		t.Fatalf("error in calling flush :: %v", err)
	}
	f1, errFile := os.OpenFile(testPath, os.O_RDWR, 0644)
	if errFile != nil {
		t.Fatalf("error in opening file :: %v", errFile)
	}
	fileData, errFile := ioutil.ReadAll(f1)
	if errFile != nil {
		t.Fatalf("error in reading file :: %s", errFile)
	}
	if err := f1.Close(); err != nil {
		t.Fatalf("error in closing file :: %v", err)
	}
	if !bytes.Equal(fileData, []byte("012345678aABCDEFGHIJKLMNOPQRSTUVWXYZ")) {
		t.Fatalf("no modification in file :: %v", string(fileData))
	}

	// Write slice bigger than mapped region after offset
	_ = m.WriteStringAt("abc", 34)
	if err := m.Flush(syscall.MS_SYNC); err != nil {
		t.Fatalf("error in flushing mapped region :: %v", err)
	}
	f2, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("error in opening file :: %v", err)
	}
	fileData, err = ioutil.ReadAll(f2)
	if err != nil {
		t.Fatalf("error in reading file :: %s", err)
	}
	if err := f2.Close(); err != nil {
		t.Fatalf("error in closing file :: %s", err)
	}
	if !bytes.Equal(fileData, []byte("012345678aABCDEFGHIJKLMNOPQRSTUVWXab")) {
		t.Fatalf("no modification in file :: %v", string(fileData))
	}

	// Write offset larger than size of mapped region
	func() {
		defer func() {
			if err := recover(); err != ErrIndexOutOfBound {
				t.Fatalf("unexpected error in writing to mmaped region :: %v", err)
			}
		}()

		_ = m.WriteStringAt("a", 100)
	}()
}

func TestAdvise(t *testing.T) {
	setup(t)
	defer tearDown(t)

	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("error in opening file :: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("error in closing file :: %v", err)
		}
	}()

	m, err := NewSharedFileMmap(f, 0, len(testData), protPage)
	if err != nil {
		t.Fatalf("error in mapping :: %v", err)
	}
	defer func() {
		if err := m.Unmap(); err != nil {
			t.Fatalf("error in calling unmap :: %v", err)
		}
	}()

	if err := m.Advise(syscall.MADV_SEQUENTIAL); err != nil {
		t.Fatalf("error in calling advise :: %v", err)
	}
}

func TestLockUnlock(t *testing.T) {
	setup(t)
	defer tearDown(t)

	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("error in opening file :: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("error in closing file :: %v", err)
		}
	}()

	m, err := NewSharedFileMmap(f, 0, len(testData), protPage)
	if err != nil {
		t.Fatalf("error in mapping :: %v", err)
	}
	defer func() {
		if err := m.Unmap(); err != nil {
			t.Fatalf("error in calling unmap :: %v", err)
		}
	}()

	if err := m.Lock(); err != nil {
		t.Fatalf("error in calling lock on mmap region :: %v", err)
	}
	if err := m.Unlock(); err != nil {
		t.Fatalf("error in calling unlock on mmap region :: %v", err)
	}
}

func TestReadUint64At(t *testing.T) {
	setup(t)
	defer tearDown(t)

	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("error in opening file :: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("error in closing file :: %v", err)
		}
	}()

	m, err := NewSharedFileMmap(f, 0, len(testData), protPage)
	if err != nil {
		t.Fatalf("error in mapping :: %v", err)
	}
	defer func() {
		if err := m.Unmap(); err != nil {
			t.Fatalf("error in calling unmap :: %v", err)
		}
	}()

	data := []byte{0x00, 0xe4, 0x0b, 0x54, 0x02, 0x00, 0x00, 0x00}

	if b, err := m.WriteAt(data, 0); err != nil {
		t.Fatalf("error in write :: %v", err)
	} else if b != 8 {
		t.Fatalf("error in writing, exp: 8, actual: %v", b)
	}

	expectedNum := uint64(10000000000)
	actualNum := m.ReadUint64At(0)
	if expectedNum != actualNum {
		t.Fatalf("Error in ReadUint64At, expected: %d, actual: %d", expectedNum, actualNum)
	}
}

func TestWriteUint64At(t *testing.T) {
	setup(t)
	defer tearDown(t)

	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("error in opening file :: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("error in closing file :: %v", err)
		}
	}()

	m, err := NewSharedFileMmap(f, 0, len(testData), protPage)
	if err != nil {
		t.Fatalf("error in mapping :: %v", err)
	}
	defer func() {
		if err := m.Unmap(); err != nil {
			t.Fatalf("error in calling unmap :: %v", err)
		}
	}()

	num := uint64(10000000000)
	m.WriteUint64At(num, 0)

	expectedSlice := []byte{0x00, 0xe4, 0x0b, 0x54, 0x02, 0x00, 0x00, 0x00}
	actualSlice := make([]byte, 8)

	if n, err := m.ReadAt(actualSlice, 0); err != nil {
		t.Fatalf("error in reading :: %v", err)
	} else if n != 8 {
		t.Fatalf("error in reading, exp: 8, actual: %v", n)
	}
	if !bytes.Equal(expectedSlice, actualSlice) {
		t.Fatalf("error in TestWriteUint64At, expected: %v, actual: %v",
			expectedSlice, actualSlice)
	}
}

func TestFailScenarios(t *testing.T) {
	setup(t)
	defer tearDown(t)

	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("error in opening file :: %v", err)
	}

	if _, err := NewSharedFileMmap(f, 0, -1, protPage); err == nil {
		t.Fatal("expected error, but found no error")
	}

	m, err := NewSharedFileMmap(f, 0, len(testData), protPage)
	if err != nil {
		t.Fatalf("error in mapping :: %v", err)
	}

	// should be safe to close here
	if err := f.Close(); err != nil {
		t.Fatalf("error in closing file :: %v", err)
	}

	num := uint64(10000000000)
	func() {
		defer func() {
			if err := recover(); err != ErrIndexOutOfBound {
				t.Fatalf("different error than expected in WriteUint64At :: %v", err)
			}
		}()

		m.WriteUint64At(num, int64(len(testData)-4))
	}()
	func() {
		defer func() {
			if err := recover(); err != ErrIndexOutOfBound {
				t.Fatalf("different error than expected in WriteUint64At :: %v", err)
			}
		}()

		_ = m.ReadUint64At(int64(len(testData) - 7))
	}()

	m.WriteUint64At(num, 0)
	if err := m.Unmap(); err != nil {
		t.Fatalf("error in calling unmap :: %v", err)
	}

	buf := make([]byte, 8)
	func() {
		defer func() {
			if err := recover(); err != ErrUnmappedMemory {
				t.Fatalf("different error than expected in ReadUint64At :: %v", err)
			}
		}()

		_ = m.ReadUint64At(0)
	}()
	func() {
		defer func() {
			if err := recover(); err != ErrUnmappedMemory {
				t.Fatalf("different error than expected in WriteUint64At :: %v", err)
			}
		}()

		m.WriteUint64At(num, 0)
	}()

	func() {
		defer func() {
			if err := recover(); err != ErrUnmappedMemory {
				t.Fatalf("different error than expected in ReadAt :: %v", err)
			}
		}()

		_, _ = m.ReadAt(buf, 0)
	}()
	func() {
		defer func() {
			if err := recover(); err != ErrUnmappedMemory {
				t.Fatalf("different error than expected in WriteAt :: %v", err)
			}
		}()

		_, _ = m.WriteAt(buf, 0)
	}()
}

func TestIOInterfaces(t *testing.T) {
	fileName := "/tmp/io.zip"
	zipData := append([]byte{80, 75, 05, 06}, bytes.Repeat([]byte{0}, 18)...)
	fileSize := len(zipData)

	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("error in creating file :: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("error in closing file :: %v", err)
		}
	}()
	defer func() {
		if err := os.Remove(fileName); err != nil {
			t.Fatalf("error in removing file :: %v", err)
		}
	}()

	m, err := NewSharedFileMmap(f, 0, fileSize, protPage)
	if err != nil {
		t.Fatalf("error in mapping :: %v", err)
	}
	defer func() {
		if err := m.Unmap(); err != nil {
			t.Fatalf("error in calling unmap :: %v", err)
		}
	}()

	if err := os.Truncate(fileName, int64(fileSize)); err != nil {
		t.Fatalf("error in truncating file :: %v", err)
	}

	if _, err := m.WriteAt(zipData, 0); err != nil {
		t.Fatalf("error in writing to file :: %v", err)
	}
	if err := m.Flush(syscall.MS_SYNC); err != nil {
		t.Fatalf("error in calling sync :: %v", err)
	}

	zr, err := zip.NewReader(m, int64(fileSize))
	if err != nil {
		t.Fatalf("error in calling unmap :: %v", err)
	}
	if len(zr.File) != 0 {
		t.Fatalf("zip file not read correctly, contains unexpected file")
	}
}

func TestDirtyFlag(t *testing.T) {
	setup(t)
	defer tearDown(t)

	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("error in opening file :: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("error in closing file :: %v", err)
		}
	}()

	m, err := NewSharedFileMmap(f, 0, len(testData), protPage)
	if err != nil {
		t.Fatalf("error in mapping :: %v", err)
	}
	defer func() {
		if err := m.Unmap(); err != nil {
			t.Fatalf("error in calling unmap :: %v", err)
		}
	}()

	m.WriteUint64At(0, 0)
	if !m.dirty {
		t.Fatalf("expected file to be dirty")
	}
	if err := m.Flush(syscall.MS_SYNC); err != nil {
		t.Fatalf("error in calling flush :: %v", err)
	}
	if m.dirty {
		t.Fatalf("expected file to be not dirty")
	}
	_ = m.ReadUint64At(0)
	// No actual syscall here
	if err := m.Flush(syscall.MS_SYNC); err != nil {
		t.Fatalf("error in calling flush :: %v", err)
	}
	if m.dirty {
		t.Fatalf("expected file to be not dirty")
	}

	_ = m.WriteStringAt("string", 0)
	if !m.dirty {
		t.Fatalf("expected file to be dirty")
	}
	if err := m.Flush(syscall.MS_SYNC); err != nil {
		t.Fatalf("error in calling flush :: %v", err)
	}
	if m.dirty {
		t.Fatalf("expected file to be not dirty")
	}
	sb := &strings.Builder{}
	sb.Grow(len("string"))
	_ = m.ReadStringAt(sb, 0)
	if m.dirty {
		t.Fatalf("expected file to be not dirty")
	}

	_, _ = m.WriteAt([]byte{1, 2}, 0)
	if !m.dirty {
		t.Fatalf("expected file to be dirty")
	}
	if err := m.Flush(syscall.MS_SYNC); err != nil {
		t.Fatalf("error in calling flush :: %v", err)
	}
	if m.dirty {
		t.Fatalf("expected file to be not dirty")
	}
	bs := make([]byte, 2)
	_, _ = m.ReadAt(bs, 0)
	if m.dirty {
		t.Fatalf("expected file to be not dirty")
	}
}
