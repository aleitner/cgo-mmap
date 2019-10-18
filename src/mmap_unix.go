package cgommap

// #cgo CFLAGS: -g -Wall
// #include <sys/mman.h>
import "C"
import (
	"errors"
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

// Mmap creates a memory map of a file
func Mmap(length, offset int64, prot, flags, fd int) (uintptr, error) {
	var cprot C.int

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

	return uintptr(C.mmap(C.NULL, C.size_t(length), C.int(cprot), C.int(flags), C.int(fd), C.longlong(offset))), nil
}

// Munmap deletes the mappings for the specified address range
func Munmap(address uintptr, length int64) error {
	success := int(C.munmap(unsafe.Pointer(address), C.ulong(length)))
	if success == -1 {
		// TODO: Check errno
		errors.New("Failed to unmap")
	}

	return nil
}
