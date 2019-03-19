package mmap

import (
	"syscall"
	"unsafe"
)

// Advise provides hints to kernel regarding the use of memory mapped region
func (m *mmapFile) Advise(advice int) error {
	_, _, err := syscall.Syscall(syscall.SYS_MADVISE,
		uintptr(unsafe.Pointer(&m.data[0])), uintptr(m.length), uintptr(advice))
	if err != 0 {
		return err
	}

	return nil
}

// Lock locks all the mapped memory to RAM, preventing the pages from swapping out
func (m *mmapFile) Lock() error {
	_, _, err := syscall.Syscall(syscall.SYS_MLOCK,
		uintptr(unsafe.Pointer(&m.data[0])), uintptr(m.length), 0)
	if err != 0 {
		return err
	}

	return nil
}

// Unlock unlocks the mapped memory from RAM, enabling swapping out of RAM if required
func (m *mmapFile) Unlock() error {
	_, _, err := syscall.Syscall(syscall.SYS_MUNLOCK,
		uintptr(unsafe.Pointer(&m.data[0])), uintptr(m.length), 0)
	if err != 0 {
		return err
	}

	return nil
}
