package mmap

import (
	"errors"
	"io"
	"os"
	"strings"
	"syscall"
)

var (
	// ErrUnmappedMemory is returned when a function is called on unmapped memory
	ErrUnmappedMemory = errors.New("unmapped memory")
	// ErrIndexOutOfBound is returned when given offset lies beyond the mapped region
	ErrIndexOutOfBound = errors.New("offset out of mapped region")
)

// File provides an interface to a memory mapped file
type File interface {
	io.ReaderAt
	io.WriterAt

	ReadStringAt(dest *strings.Builder, offset int64) int
	WriteStringAt(src string, offset int64) int
	ReadUint64At(offset int64) uint64
	WriteUint64At(num uint64, offset int64)

	Lock() error
	Unlock() error
	Advise(advice int) error
	Flush(flags int) error
	Unmap() error
}

// mmapFile provides abstraction around a memory mapped file
type mmapFile struct {
	data   []byte
	length int64
}

// NewSharedFileMmap maps a file into memory starting at a given offset, for given length.
// For documentation regarding prot, see documentation for syscall package.
// possible cases:
//    case 1 => if   file size > memory region (offset + length)
//              then all the mapped memory is accessible
//    case 2 => if   file size <= memory region (offset + length)
//              then from offset to file size memory region is accessible
func NewSharedFileMmap(f *os.File, offset int64, length int, prot int) (File, error) {
	data, err := syscall.Mmap(int(f.Fd()), offset, length, prot, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	return &mmapFile{
		data:   data,
		length: int64(length),
	}, nil
}

// Unmap unmaps the memory mapped file. An error will be returned
// if any of the functions are called on Mmap after calling Unmap
func (m *mmapFile) Unmap() error {
	err := syscall.Munmap(m.data)
	m.data = nil
	return err
}
