package mmap

import (
	"errors"
	"os"
	"syscall"
)

var (
	// ErrUnmappedMemory is returned when a function is called on unmapped memory
	ErrUnmappedMemory = errors.New("unmapped memory")
)

// Mmap provides abstraction around a memory mapped file
type Mmap struct {
	data   []byte
	length int
}

// NewMmap maps a file into memory starting at a given offset, for given length.
// For documentation regarding prot and flags, see documentation for syscall package.
// possible cases:
//    case 1 => if   file size > memory region (offset + length)
//              then all the mapped memory is accessible
//    case 2 => if   file size <= memory region (offset + length)
//              then from offset to file size memory region is accessible
func NewMmap(f *os.File, offset int64, length int, prot int, flags int) (*Mmap, error) {
	data, err := syscall.Mmap(int(f.Fd()), offset, length, prot, flags)
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

// Read copies <size> bytes starting at offset (inclusive) to destination slice.
// Caller needs to ensure that length of dest slice > size
func (m *Mmap) Read(dest []byte, offset int, size int) error {
	if m.data == nil {
		return ErrUnmappedMemory
	}

	copy(dest, m.data[offset:offset+size])
	return nil
}

// Write copies the whole given slice at offset in the memory mapped file/region
func (m *Mmap) Write(src []byte, offset int) error {
	if m.data == nil {
		return ErrUnmappedMemory
	}

	copy(m.data[offset:], src)
	return nil
}
