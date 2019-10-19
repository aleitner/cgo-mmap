package cgommap

// #cgo CFLAGS: -g -Wall
// #include <string.h>
import "C"
import (
	"errors"
	"unsafe"
)

const (

	// The prot argument describes the desired memory protection of the
	// mapping (and must not conflict with the open mode of the file).
	PROT_EXEC_READ = iota +1 // Allows views to be mapped for read-only, copy-on-write, or execute access.
	PROT_EXEC_WRITE          // Allows views to be mapped for read-only, copy-on-write, or execute access.
	PROT_EXEC_READWRITE      // Allows views to be mapped for read-only, copy-on-write, read/write, or execute access.
	PROT_READ        	     // Pages may be read.
	PROT_WRITE               // Pages may be written.
	PROT_READWRITE           // Pages may not be accessed.
)

// MMAP contains information about a memory mapped file
type MMAP struct {
	addr uintptr
	size int64
	offset int64
}

// NewMmap opens a memory mapped file.
// TODO: Check if go bitwise `|` operator works on C.int properly
func NewMmap(length, offset int64, prot, flags, fd int) (mmap *MMAP, err error) {
	if offset > length || offset < 0 {
		return nil, errors.New("Invalid offset")
	}
	address, err := Mmap(length, offset, prot, flags, fd)
	if err != nil {
		return nil, err
	}

	return &MMAP{
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
	buf = (*[1 << 30]byte)(unsafe.Pointer(mmap.addr+uintptr(mmap.offset)))[:mmap.size-mmap.offset]
	mmap.offset += int64(len(buf))

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
	return Munmap(mmap.addr, mmap.size)
}
