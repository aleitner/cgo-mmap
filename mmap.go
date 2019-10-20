package cgommap

// #cgo CFLAGS: -g -Wall
// #include <string.h>
import "C"
import (
	"errors"
	"io"
	"sync"
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
	mtx sync.Mutex
}

// NewMmap opens a memory mapped file.
// TODO: Check if go bitwise `|` operator works on C.int properly
func NewMmap(length, offset int64, prot, flags int, fd uintptr) (mmap *MMAP, err error) {
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
func (mmap *MMAP) Write(buf []byte) (writeLen int, err error) {
	writeLen = len(buf)

	// NB: Should we just return instead of partial Write?
	if int64(writeLen) > mmap.size - mmap.offset {
		writeLen = int(mmap.size - mmap.offset)
		if writeLen == 0 {
			return 0, io.EOF
		}
		err = errors.New("Partial Write")
	}

	C.memcpy(unsafe.Pointer(mmap.addr + uintptr(mmap.offset)), unsafe.Pointer(&buf[0]), C.size_t(writeLen))

	mmap.mtx.Lock()
	defer mmap.mtx.Unlock()
	mmap.offset += int64(writeLen)

	return writeLen, err
}

// Read from mmap into buf
func (mmap *MMAP) Read(buf []byte) (n int, err error) {
	toReadLen := mmap.size-mmap.offset

	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{mmap.addr+uintptr(mmap.offset), 0, int(toReadLen)}

	mmapBuf := *(*[]byte)(unsafe.Pointer(&sl))

	copy(buf, mmapBuf[:toReadLen])

	if toReadLen == 0 {
		return 0, io.EOF
	}

	mmap.mtx.Lock()
	defer mmap.mtx.Unlock()
	mmap.offset += int64(toReadLen)

	return int(toReadLen), err
}

// Seek mmap
func (mmap *MMAP) Seek(off int64, origin int) (newOffset int64, err error) {
	switch origin {
	case io.SeekCurrent:
		newOffset = off + mmap.offset
		break
	case io.SeekStart:
		newOffset = off
		break
	case io.SeekEnd:
		newOffset = off + mmap.size - 1
		break
	default:
		return mmap.offset, errors.New("Invalid origin")
	}

	if newOffset > mmap.size {
		return mmap.offset, errors.New("Invalid Seek")
	}

	mmap.mtx.Lock()
	defer mmap.mtx.Unlock()
	mmap.offset = newOffset

	return newOffset, nil
}

// Size of the mmap
func (mmap *MMAP) Size() int64 {
	return mmap.size
}

// Close mmap
func (mmap *MMAP) Close() error {
	return Munmap(mmap.addr, mmap.size)
}

