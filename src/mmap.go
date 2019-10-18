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
	// mapping (and must not conflict with the open mode of the file).
	PROT_EXEC_READ = iota // Allows views to be mapped for read-only, copy-on-write, or execute access.
	PROT_EXEC_WRITE       // Allows views to be mapped for read-only, copy-on-write, or execute access.
	PROT_EXEC_READWRITE   // Allows views to be mapped for read-only, copy-on-write, read/write, or execute access.
	PROT_READ        	  // Pages may be read.
	PROT_WRITE            // Pages may be written.
	PROT_READWRITE        // Pages may not be accessed.

	// The flags argument determines whether updates to the mapping are
	// visible to other processes mapping the same region, and whether
	// updates are carried through to the underlying file.  This behavior is
	// determined by including exactly one of the following values in flags:
	MAP_SHARED          // Share this mapping.
	MAP_SHARED_VALIDATE // This flag provides the same behavior as MAP_SHARED except that MAP_SHARED mappings ignore unknown flags in flags.
	MAP_PRIVATE         // Create a private copy-on-write mapping.
	// TODO: Add other mappings
)

// MMAP contains information about a memmory mapped file
type MMAP struct {
	file *File
	addr uintptr
	size int64
	offset int64
}

// NewMmap opens a memory mapped file.
// TODO: Check if go bitwise `|` operator works on C.int properly
func NewMmap(length, offset int64, prot, flags C.int, filepath string, mode int) (mmap *MMAP, err error) {
	f := OpenFile(filepath, mode)

	defer func() {
		if err != nil {
			f.Close()
		}
	}()

	fd, err := f.Fileno()
	if err != nil {
		return nil, err
	}

	address, err := Mmap(length, offset, prot, flags, fd)
	if err != nil {
		return nil, err
	}

	return &MMAP{
		file: f,
		addr: address,
		size: length,
	}, nil
}

// Write buf to mmap
func (mmap *MMAP) Write(buf []byte) (writeLen int64, err error) {
	writeLen = int64(len(buf))

	// NB: Should we just return instead of partial Write?
	if writeLen > mmap.size - mmap.offset {
		writeLen = mmap.size - mmap.offset
		err = errors.New("Partial Write")
	}

	// TODO: Does this math work when moving around the memory?
	C.memcpy(unsafe.Pointer(mmap.addr + uintptr(mmap.offset)), unsafe.Pointer(&buf[0]), C.size_t(writeLen))
	mmap.offset += writeLen

	return writeLen, err
}

// Read from mmap into buf
func (mmap *MMAP) Read(buf []byte) (int, error) {
	buf = (*[1 << 30]byte)(unsafe.Pointer(mmap.addr+uintptr(mmap.offset)))[:cap(buf)]

	//NB: Should we return EOF or an error if len(buf) == 0?
	return len(buf), nil
}

// Seek mmap
func (mmap *MMAP) Seek(off int64, origin int) (newOffset int64, err error) {
	switch origin {
	case SEEK_CUR:
		newOffset = off + mmap.offset
		break
	case SEEK_SET:
		newOffset = off
		break
	case SEEK_END:
		newOffset = off + mmap.size - 1
		break
	default:
		return mmap.offset, errors.New("Invalid origin")
	}

	if newOffset > mmap.size {
		return mmap.offset, errors.New("Invalid Seek")
	}
	mmap.offset = newOffset

	return newOffset, nil
}


// Close mmap
func (mmap *MMAP) Close() error {
	// TODO: Handle error
	Munmap(mmap.addr, mmap.size)

	// TODO: Handle error
	mmap.file.Close()

	return nil
}

