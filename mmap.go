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

// New opens a memory mapped file.
func New(length, offset int64, prot, flags int, fd uintptr) (m *MMAP, err error) {
	if offset > length || offset < 0 {
		return nil, errors.New("Invalid offset")
	}
	address, err := mmap(length, offset, prot, flags, fd)
	if err != nil {
		return nil, err
	}

	return &MMAP{
		addr: address,
		size: length,
	}, nil
}

// Write buf to mmap
func (m *MMAP) Write(buf []byte) (writeLen int, err error) {
	writeLen = len(buf)

	// NB: Should we just return instead of partial Write?
	if int64(writeLen) > m.size - m.offset {
		writeLen = int(m.size - m.offset)
		if writeLen == 0 {
			return 0, io.EOF
		}
		err = errors.New("Partial Write")
	}

	C.memcpy(unsafe.Pointer(m.addr + uintptr(m.offset)), unsafe.Pointer(&buf[0]), C.size_t(writeLen))

	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.offset += int64(writeLen)

	return writeLen, err
}

// Read from mmap into buf
func (m *MMAP) Read(buf []byte) (n int, err error) {
	toReadLen := m.size-m.offset

	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{m.addr+uintptr(m.offset), 0, int(toReadLen)}

	mmapBuf := *(*[]byte)(unsafe.Pointer(&sl))

	copy(buf, mmapBuf[:toReadLen])

	if toReadLen == 0 {
		return 0, io.EOF
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.offset += int64(toReadLen)

	return int(toReadLen), err
}

// Seek mmap
func (m *MMAP) Seek(off int64, origin int) (newOffset int64, err error) {
	switch origin {
	case io.SeekCurrent:
		newOffset = off + m.offset
		break
	case io.SeekStart:
		newOffset = off
		break
	case io.SeekEnd:
		newOffset = off + m.size - 1
		break
	default:
		return m.offset, errors.New("Invalid origin")
	}

	if newOffset > m.size {
		return m.offset, errors.New("Invalid Seek")
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.offset = newOffset

	return newOffset, nil
}

// Size of the mmap
func (m MMAP) Size() int64 {
	return m.size
}

// Close mmap
func (m MMAP) Close() error {
	return munmap(m.addr, m.size)
}

// TODO

func (m MMAP) lock() error {
	return nil
}

func (m MMAP) unlock() error {
	return nil
}

func (m MMAP) Flush() error {
	return nil
}