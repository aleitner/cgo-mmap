package cgommap

// #cgo CFLAGS: -g -Wall
// #include <windows.h>
import "C"
import "unsafe"

// mmap creates a memory map of a file
func mmap(length, offset int64, prot, flags int, fd uintptr) (uintptr, error) {
	fh := C.HANDLE(C._get_osfhandle(C.int(fd)))
	if (fh == C.INVALID_HANDLE_VALUE) {
		return 0, errors.New("mmap error: Handle error: Invalid Handle Value")
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
		return 0, errors.New("mmap error: Invalid protection value")
	}

	mh := C.CreateFileMapping(fh, C.NULL, C.int(protection), C.int(0), C.int(0), C.NULL)
	if (!mh) {
		return 0, os.NewSyscallError("mmap error: CreateFileMapping error", C.GetLastError())
	}

	defer C.CloseHandle(mh)

	return C.MapViewOfFileEx(mh, C.int(desiredAccess), C.int(0), C.int(0), C.int(length), C.NULL), nil
}

// unmap deletes the mappings for the specified address range
func unmap(addr uintptr, length int64) error {

	// TODO: Handle errorss
	_ = flush(addr, length)
	_ = C.UnmapViewOfFile(unsafe.Pointer(addr))
}

// lock the mapped memory, ensuring that subsequent access to the region will not incur a page fault.
func lock(addr uintptr, len int64) error {
	if success := C.VirtualLock(unsafe.Pointer(addr), C.SIZE_T(len)); success != 0 {
		return os.NewSyscallError("lock error: VirtualLock error", C.GetLastError())
	}
	return nil
}

// unlock the mapped memory, enabling the system to swap the pages out to the paging file if necessary.
func unlock(addr uintptr, len int64) error {
	if success := C.VirtualUnlock(unsafe.Pointer(addr), C.SIZE_T(len)); success != 0 {
		return os.NewSyscallError("unlock error: VirtualUnlock error", C.GetLastError())
	}
	return nil
}

// flush the mapped view of a file to disk.
func flush(addr uintptr, length int64, flags int) error {
	if success := C.FlushViewOfFile(unsafe.Pointer(addr), C.int(len)); success != 0 {
		return os.NewSyscallError("flush error: FlushViewOfFile error", C.GetLastError())
	}

	return nil
}