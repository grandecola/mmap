package mmap

import "io"

// IMmap provides an interface to a memory mapped file
type IMmap interface {
	io.ReaderAt
	io.WriterAt

	Lock() error
	Unlock() error
	Advise(advice int) error
	ReadUint64At(offset int64) (uint64, error)
	WriteUint64At(num uint64, offset int64) error
	Flush(flags int) error
	Unmap() error
}
