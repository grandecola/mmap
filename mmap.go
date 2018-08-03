package mmap

import (
	"encoding/binary"
	"errors"
	"os"
	"syscall"
)

var (
	// ErrUnmappedMemory is returned when a function is called on unmapped memory
	ErrUnmappedMemory = errors.New("unmapped memory")
	// ErrIndexOutOfBound is returned when given offset lies beyond the mapped region
	ErrIndexOutOfBound = errors.New("offset out of mapped region")
)

// Mmap provides abstraction around a memory mapped file
type Mmap struct {
	data   []byte
	length int
}

// NewSharedFileMmap maps a file into memory starting at a given offset, for given length.
// For documentation regarding prot, see documentation for syscall package.
// possible cases:
//    case 1 => if   file size > memory region (offset + length)
//              then all the mapped memory is accessible
//    case 2 => if   file size <= memory region (offset + length)
//              then from offset to file size memory region is accessible
func NewSharedFileMmap(f *os.File, offset int64, length int, prot int) (*Mmap, error) {
	data, err := syscall.Mmap(int(f.Fd()), offset, length, prot, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	return &Mmap{
		data:   data,
		length: length,
	}, nil
}

// Unmap unmaps the memory mapped file. An error will be returned
// if any of the functions are called on Mmap after calling Unmap
func (m *Mmap) Unmap() error {
	err := syscall.Munmap(m.data)
	m.data = nil
	return err
}

// Read copies data to dest slice from mapped region starting at given offset
func (m *Mmap) Read(dest []byte, offset int) (int, error) {
	if m.data == nil {
		return 0, ErrUnmappedMemory
	} else if offset >= m.length || offset < 0 {
		return 0, ErrIndexOutOfBound
	}

	return copy(dest, m.data[offset:]), nil
}

// Write copies data to mapped region from the src slice starting at given offset
func (m *Mmap) Write(src []byte, offset int) (int, error) {
	if m.data == nil {
		return 0, ErrUnmappedMemory
	} else if offset >= m.length || offset < 0 {
		return 0, ErrIndexOutOfBound
	}

	return copy(m.data[offset:], src), nil
}

// ReadUint64 reads uint64 from offset
func (m *Mmap) ReadUint64(offset int) uint64 {
	return binary.LittleEndian.Uint64(m.data[offset : offset+8])
}

// WriteUint64 writes num at offset
func (m *Mmap) WriteUint64(offset int, num uint64) {
	binary.LittleEndian.PutUint64(m.data[offset:offset+8], num)
}
