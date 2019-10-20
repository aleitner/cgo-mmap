package cgommap

// #cgo CFLAGS: -g -Wall
// #include <sys/mman.h>
// #include <stdio.h>
// #include <stdlib.h>
// #include <unistd.h>
// #include <errno.h>
// #include <string.h>
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

const (
	MS_SYNC = iota // Requests an update and waits for it to complete.
	MS_ASYNC // Specifies that an update be scheduled, but the call returns immediately.
    MS_INVALIDATE // Asks to invalidate other mappings of the same file

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

// unmap deletes the mappings for the specified address range
func unmap(addr uintptr, length int64) error {
	// TODO: handle multiple errors
	if err := flush(addr, length, MS_SYNC); err != nil {
		return err
	}

	if success := int(C.munmap(unsafe.Pointer(addr), C.ulong(length))); success != 0 {
		return fmt.Errorf("unmap error: munmap error: %s", C.GoString(C.strerror(C.int(success))))
	}

	return nil
}

// lock the calling process's virtual address space into RAM, preventing that memory from being paged to the swap area.
func lock(addr uintptr, length int64) error {
	if success := C.mlock(unsafe.Pointer(addr), C.size_t(length)); success != 0 {
		return fmt.Errorf("lock error: mlock error: %s", C.GoString(C.strerror(C.int(success))))
	}

	return nil
}

// unlock the calling process's virtual address space, so that
// pages in the specified virtual address range may once more to be
// swapped out if required by the kernel memory manager.
func unlock(addr uintptr, length int64) error {
	if success := C.munlock(unsafe.Pointer(addr), C.size_t(length)); success != 0 {
		return fmt.Errorf("unlock error: munlock error: %s", C.GoString(C.strerror(C.int(success))))
	}

	return nil
}

// flush changes made to the memory map back to the filesystem
func flush(addr uintptr, length int64, flags int) error {
	if success := C.msync(unsafe.Pointer(addr), C.size_t(length), C.int(flags)); success != 0 {
		return fmt.Errorf("flush error: msync error: %s", C.GoString(C.strerror(C.int(success))))
	}

	return nil
}