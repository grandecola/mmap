package mmap

import (
	"encoding/binary"
	"syscall"
	"unsafe"
)

// Read copies data to dest slice from mapped region starting at
// given offset and returns number of bytes copied to the dest slice.
// There are two possibilities -
//   Case 1: len(dest) >= (len(m.data) - offset)
//        => copies (len(m.data) - offset) bytes to dest from mapped region
//   Case 2: len(dest) < (len(m.data) - offset)
//        => copies len(dest) bytes to dest from mapped region
func (m *Mmap) Read(dest []byte, offset int) (int, error) {
	if m.data == nil {
		return 0, ErrUnmappedMemory
	} else if offset >= m.length || offset < 0 {
		return 0, ErrIndexOutOfBound
	}

	return copy(dest, m.data[offset:]), nil
}

// Write copies data to mapped region from the src slice starting at
// given offset and returns number of bytes copied to the mapped region.
// There are two possibilities -
//  Case 1: len(src) >= (len(m.data) - offset)
//      => copies (len(m.data) - offset) bytes to the mapped region from src
//  Case 2: len(src) < (len(m.data) - offset)
//      => copies len(src) bytes to the mapped region from src
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

// Flush flushes the memory mapped region to disk
func (m *Mmap) Flush(flags int) error {
	_, _, err := syscall.Syscall(syscall.SYS_MSYNC,
		uintptr(unsafe.Pointer(&m.data[0])), uintptr(len(m.data)), uintptr(flags))
	if err != 0 {
		return err
	}

	return nil
}
