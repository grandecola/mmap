# mmap [![Build Status](https://travis-ci.com/grandecola/mmap.svg?branch=master)](https://travis-ci.com/grandecola/mmap) [![Go Report Card](https://goreportcard.com/badge/github.com/grandecola/mmap)](https://goreportcard.com/report/github.com/grandecola/mmap) [![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT) [![GoDoc](https://godoc.org/github.com/grandecola/mmap?status.svg)](https://godoc.org/github.com/grandecola/mmap) [![codecov](https://codecov.io/gh/grandecola/mmap/branch/master/graph/badge.svg)](https://codecov.io/gh/grandecola/mmap)
Interface for mmap syscall to provide safe and efficient access to memory.
`*mmap.File` satisfies both `io.ReaderAt` and `io.WriterAt` interfaces.

**Only works for darwin OS, Linux and Little Endian 64 bit architectures.**

## Safety & Efficiency
Golang mmap syscall function exposes the mapped memory as array of bytes.
If the array is referenced even after the memory region is unmapped,
this can lead to segmentation fault. `mmap package` provides safe access
to the array of bytes by providing `ReadAt` and `WriteAt` functions.

`WriteAt` function copies a slice into the memory mapped region
whereas `ReadAt` function copies data from memory mapped region to
a given slice, therefore, avoiding exposing the array of bytes referring
to mapped memory. This also avoids any extra data copy providing efficient
access to the memory mapped region.

We have also added functions such as `WriteUint64At`, `ReadUint64At` that
can directly typecast the mmaped memory to Uint64 and avoids an extra copy.
We will add more functions in the library based on our use cases. If you need
support for a particular function, let us know or better, raise a pull request.

## Similar Packages
* golang.org/x/exp/mmap
* github.com/riobard/go-mmap
* launchpad.net/gommap
* github.com/edsrzf/mmap-go
