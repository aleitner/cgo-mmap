package cgommap

// #cgo CFLAGS: -g -Wall
// #include <windows.h>
import "C"

// Mmap creates a memory map of a file
func Mmap(length, offset int64, prot, flags, fd int) (uintptr, error) {
	fh := C.HANDLE(C._get_osfhandle(C.int(fd)))
	if (fh == C.INVALID_HANDLE_VALUE) {
		return 0, errors.New("Invalid Handle Value")
	}

	var protection C.int
	var desiredAccess C.int

	switch(prot) {
	case PROT_READ:
		protection = C.PAGE_READONLY
		desiredAccess = C.FILE_MAP_READ
		break
	case PROT_WRITE:
		protection = C.PAGE_WRITECOPY
		desiredAccess = C.FILE_MAP_WRITE | C.FILE_MAP_COPY
		break
	case PROT_READWRITE:
		protection = C.PAGE_READWRITE
		desiredAccess = C.FILE_MAP_ALL_ACCESS
		break
	case PROT_EXEC_READ:
		protection = C.PAGE_EXECUTE_READ
		desiredAccess = C.FILE_MAP_EXECUTE| C.FILE_MAP_READ
		break
	case PROT_EXEC_WRITE:
		protection = C.PAGE_EXECUTE_WRITECOPY
		desiredAccess = C.FILE_MAP_EXECUTE | C.FILE_MAP_WRITE | C.FILE_MAP_COPY
		break
	case PROT_EXEC_READWRITE:
		protection = C.PAGE_EXECUTE_READWRITE
		desiredAccess = C.FILE_MAP_EXECUTE | FILE_MAP_ALL_ACCESS
		break
	default:
		return 0, errors.New("Invalid protection value")
	}

	mh := C.CreateFileMapping(fh, C.NULL, C.int(protection), C.int(0), C.int(0), C.NULL)
	if (!mh) {
		return 0, errors.New("Failed to create file mapping")
	}

	defer C.CloseHandle(mh)

	return C.MapViewOfFileEx(mh, C.int(desiredAccess), C.int(0), C.int(0), C.int(length), C.NULL), nil
}

// Munmap deletes the mappings for the specified address range
func Munmap(address uintptr, length int64) error {
	C.FlushViewOfFile(address, C.int(length))
	C.UnmapViewOfFile(address)
}
