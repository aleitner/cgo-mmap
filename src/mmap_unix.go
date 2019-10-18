package cgommap

import "C"
import "errors"

// Mmap creates a memory map of a file
func Mmap(length, offset int64, prot, flags C.int, fd int) uintptr {
	return uintptr(C.mmap(0, C.size_t(length), prot, flags, fd, C.int(offset)))
}

// Munmap deletes the mappings for the specified address range
func Munmap(address uintptr, length int64) error {
	success := int(C.munmap(address, C.int(length)))
	if success == -1 {
		// TODO: Check errno
		errors.New("Failed.")
	}

	return nil
}
