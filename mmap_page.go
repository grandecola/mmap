package mmap

import (
	"golang.org/x/sys/unix"
)

// Advise provides hints to kernel regarding the use of memory mapped region
func (m *Mmap) Advise(advice int) error {
	return unix.Madvise(m.data, advice)
}

// Lock locks all the mapped memory to RAM, preventing the pages from swapping out
func (m *Mmap) Lock() error {
	return unix.Mlock(m.data)
}

// Unlock unlocks the mapped memory from RAM, enabling swapping out of RAM if required
func (m *Mmap) Unlock() error {
	return unix.Munlock(m.data)
}
