package mmap

import (
	"encoding/binary"
	"strings"
	"syscall"
	"unsafe"
)

func (m *FileImpl) boundaryChecks(offset, numBytes int64) {
	if m.data == nil {
		panic(ErrUnmappedMemory)
	} else if offset+numBytes > m.length || offset < 0 {
		panic(ErrIndexOutOfBound)
	}
}

// ReadAt copies data to dest slice from mapped region starting at
// given offset and returns number of bytes copied to the dest slice.
// There are two possibilities -
//   Case 1: len(dest) >= (m.length - offset)
//        => copies (m.length - offset) bytes to dest from mapped region
//   Case 2: len(dest) < (m.length - offset)
//        => copies len(dest) bytes to dest from mapped region
// err is always nil, hence, can be ignored
func (m *FileImpl) ReadAt(dest []byte, offset int64) (int, error) {
	m.boundaryChecks(offset, 1)
	return copy(dest, m.data[offset:]), nil
}

// WriteAt copies data to mapped region from the src slice starting at
// given offset and returns number of bytes copied to the mapped region.
// There are two possibilities -
//  Case 1: len(src) >= (m.length - offset)
//      => copies (m.length - offset) bytes to the mapped region from src
//  Case 2: len(src) < (m.length - offset)
//      => copies len(src) bytes to the mapped region from src
// err is always nil, hence, can be ignored
func (m *FileImpl) WriteAt(src []byte, offset int64) (int, error) {
	m.boundaryChecks(offset, 1)
	return copy(m.data[offset:], src), nil
}

// ReadStringAt copies data to dest string builder from mapped region starting at
// given offset until the min value of (length - offset) or (dest.Cap() - dest.Len())
// and returns number of bytes copied to the dest slice.
func (m *FileImpl) ReadStringAt(dest *strings.Builder, offset int64) int {
	m.boundaryChecks(offset, 1)

	dataLength := m.length - offset
	emptySpace := int64(dest.Cap() - dest.Len())
	end := m.length
	if dataLength > emptySpace {
		end = offset + emptySpace
	}

	n, _ := dest.Write(m.data[offset:end])
	return n
}

// WriteStringAt copies data to mapped region from the src string starting at
// given offset and returns number of bytes copied to the mapped region.
// See github.com/grandecola/mmap/#Mmap.WriteAt for more details.
func (m *FileImpl) WriteStringAt(src string, offset int64) int {
	m.boundaryChecks(offset, 1)
	return copy(m.data[offset:], src)
}

// ReadUint64At reads uint64 from offset
func (m *FileImpl) ReadUint64At(offset int64) uint64 {
	m.boundaryChecks(offset, 8)
	return binary.LittleEndian.Uint64(m.data[offset : offset+8])
}

// WriteUint64At writes num at offset
func (m *FileImpl) WriteUint64At(num uint64, offset int64) {
	m.boundaryChecks(offset, 8)
	binary.LittleEndian.PutUint64(m.data[offset:offset+8], num)
}

// Flush flushes the memory mapped region to disk
func (m *FileImpl) Flush(flags int) error {
	_, _, err := syscall.Syscall(syscall.SYS_MSYNC,
		uintptr(unsafe.Pointer(&m.data[0])), uintptr(m.length), uintptr(flags))
	if err != 0 {
		return err
	}

	return nil
}
