package mmap

import (
	"encoding/binary"

	"golang.org/x/sys/unix"
)

// ReadAt copies data to dest slice from mapped region starting at
// given offset and returns number of bytes copied to the dest slice.
// There are two possibilities -
//   Case 1: len(dest) >= (len(m.data) - offset)
//        => copies (len(m.data) - offset) bytes to dest from mapped region
//   Case 2: len(dest) < (len(m.data) - offset)
//        => copies len(dest) bytes to dest from mapped region
func (m *Mmap) ReadAt(dest []byte, offset int64) (int, error) {
	if m.data == nil {
		return 0, ErrUnmappedMemory
	} else if offset >= m.length || offset < 0 {
		return 0, ErrIndexOutOfBound
	}

	return copy(dest, m.data[offset:]), nil
}

// WriteAt copies data to mapped region from the src slice starting at
// given offset and returns number of bytes copied to the mapped region.
// There are two possibilities -
//  Case 1: len(src) >= (len(m.data) - offset)
//      => copies (len(m.data) - offset) bytes to the mapped region from src
//  Case 2: len(src) < (len(m.data) - offset)
//      => copies len(src) bytes to the mapped region from src
func (m *Mmap) WriteAt(src []byte, offset int64) (int, error) {
	if m.data == nil {
		return 0, ErrUnmappedMemory
	} else if offset >= m.length || offset < 0 {
		return 0, ErrIndexOutOfBound
	}

	return copy(m.data[offset:], src), nil
}

// ReadUint64At reads uint64 from offset
func (m *Mmap) ReadUint64At(offset int64) (uint64, error) {
	if m.data == nil {
		return 0, ErrUnmappedMemory
	} else if offset+8 > m.length || offset < 0 {
		return 0, ErrIndexOutOfBound
	}

	return binary.LittleEndian.Uint64(m.data[offset : offset+8]), nil
}

// WriteUint64At writes num at offset
func (m *Mmap) WriteUint64At(num uint64, offset int64) error {
	if m.data == nil {
		return ErrUnmappedMemory
	} else if offset+8 > m.length || offset < 0 {
		return ErrIndexOutOfBound
	}

	binary.LittleEndian.PutUint64(m.data[offset:offset+8], num)
	return nil
}

// Flush flushes the memory mapped region to disk
func (m *Mmap) Flush(flags int) error {
	return unix.Msync(m.data, flags)
}
