package mmap

// IMmap provides an interface to a memory mapped file
type IMmap interface {
	Lock() error
	Unlock() error
	Advise(advice int) error
	Read(dest []byte, offset int) (int, error)
	Write(src []byte, offset int) (int, error)
	ReadUint64(offset int) uint64
	WriteUint64(offset int, num uint64)
	Flush(flags int) error
	Unmap() error
}
