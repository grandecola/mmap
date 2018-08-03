# mmap [![Build Status](https://travis-ci.com/grandecola/mmap.svg?branch=master)](https://travis-ci.com/grandecola/mmap)

Interface for mmap syscall to provide safe and efficient access to memory.

**Only works for darwin OS and Little Endian architecture.**

## Safety & Efficiency

Golang mmap syscall function exposes the mapped memory as array of bytes. If the array is referenced even after the memory region is unmapped, this can lead to segmentation fault. `mmap package` provides safe access to the array of bytes by providing `Read` and `Write` functions. `Write` function copies a slice into the memory mapped region whereas `Read` function copies data from memory mapped region to a given slice, therefore, avoiding exposing the array of bytes referring to mapped memory. This also avoids any extra data copy providing efficient access to the memory mapped region.

## Similar Packages

* golang.org/x/exp/mmap
* github.com/riobard/go-mmap
* launchpad.net/gommap
* github.com/edsrzf/mmap-go
