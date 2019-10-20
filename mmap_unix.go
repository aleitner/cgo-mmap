package cgommap
// #cgo CFLAGS: -g -Wall
// #include <sys/mman.h>
// #include <stdio.h>
// #include <stdlib.h>
// #include <unistd.h>
import "C"
import (
	"errors"
	"os"
	"unsafe"
)

const (
	// The flags argument determines whether updates to the mapping are
	// visible to other processes mapping the same region, and whether
	// updates are carried through to the underlying file.  This behavior is
	// determined by including exactly one of the following values in flags:
	MAP_SHARED = C.MAP_SHARED // Share this mapping.
	MAP_PRIVATE = C.MAP_PRIVATE // Create a private copy-on-write mapping.
	// TODO: Add other mappings
)

// mmap creates a memory map of a file
func mmap(length, offset int64, prot, flags int, fd uintptr) (uintptr, error) {
	var cprot C.int

	// TODO: Check if go bitwise `|` operator works on C.int properly
	switch(prot) {
	case PROT_READ:
		cprot = C.PROT_READ
		break
	case PROT_WRITE:
		cprot = C.PROT_WRITE
		break
	case PROT_READWRITE:
		cprot = C.PROT_READ | C.PROT_WRITE
		break
	case PROT_EXEC_READ:
		cprot = C.PROT_EXEC | C.PROT_READ
		break
	case PROT_EXEC_WRITE:
		cprot = C.PROT_EXEC | C.PROT_WRITE
		break
	case PROT_EXEC_READWRITE:
		cprot = C.PROT_EXEC | C.PROT_WRITE | C.PROT_READ
		break
	default:
		cprot = C.PROT_NONE
	}

	//NB: offset for mmap() must be page aligned.
	// get the page size and makes sure that pa_offset is either zero or a multiple of the page size (4096)
	// Assuming the page size is a power of two, then one less than the page-size will have all bits set
	// until the bit that represents the power of two corresponding to the page size.
	offset -= offset & int64(^(os.Getpagesize() - 1))

	return uintptr(C.mmap(C.NULL, C.size_t(length), C.int(cprot), C.int(flags), C.int(fd), C.longlong(offset))), nil
}

// munmap deletes the mappings for the specified address range
func munmap(address uintptr, length int64) error {
	success := int(C.munmap(unsafe.Pointer(address), C.ulong(length)))
	if success == -1 {
		// TODO: Check errno
		errors.New("Failed to unmap")
	}

	return nil
}
