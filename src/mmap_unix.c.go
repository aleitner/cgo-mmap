package cgommap

// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include <sys/mman.h>
import "C"
import (
	"errors"
	"unsafe"
)

const (

	// The prot argument describes the desired memory protection of the
	// mapping (and must not conflict with the open mode of the file).  It
	// is either PROT_NONE or the bitwise OR of one or more of the following
	// flags:
	PROT_EXEC = C.PROT_EXEC // Pages may be executed.
	PROT_READ = C.PROT_READ // Pages may be read.
	PROT_WRITE = C.PROT_WRITE // Pages may be written.
	PROT_NONE = C.PROT_NONE // Pages may not be accessed.

	// The flags argument determines whether updates to the mapping are
	// visible to other processes mapping the same region, and whether
	// updates are carried through to the underlying file.  This behavior is
	// determined by including exactly one of the following values in flags:
	MAP_SHARED = C.MAP_SHARED // Share this mapping.
	MAP_SHARED_VALIDATE = C.MAP_SHARED_VALIDATE // This flag provides the same behavior as MAP_SHARED except that MAP_SHARED mappings ignore unknown flags in flags.
	MAP_PRIVATE = C.MAP_PRIVATE // Create a private copy-on-write mapping.
	// TODO: Add other mappings
)

type MMAP struct {
	addr uintptr
	length int
	offset int
}

// mmap opens a memory mapped file.
// TODO: Check if go bitwise `|` operator works on C.int properly
func mmap(length, offset int, prot, flags C.int, filename string, mode int) (*MMAP, error) {
	f := OpenFile(filename, mode)

	fd, err := f.Fileno()
	if err != nil {
		return nil, err
	}

	address := C.mmap(0, C.size_t(length), prot, flags, fd, C.int(offset))

	return &MMAP{
		addr: address,
		length: length,
	}, nil
}

func (mmap *MMAP) Write(buf []byte) (int, error) {
	bufLen := len(buf)

	// TODO: Does this math work when moving around the memory?
	s := int(C.write(unsafe.Pointer(&buf[0]), mmap.addr + uintptr(mmap.offset), C.size_t(bufLen)))
	if (s != bufLen) {
		if (s == -1) {
			// TODO: Check errno
			return s, errors.New("Write error")
		}

		mmap.offset += s
		
		return s, errors.New("Partial Write error")
	}

	mmap.offset += s

	return s, nil
}

func munmap(address uintptr, length int) error {
	success := int(C.munmap(address, C.int(length)))
	if success == -1 {
		// TODO: Check errno
		errors.New("Failed.")
	}

	return nil
}

