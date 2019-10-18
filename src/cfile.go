package cgommap

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
*/
import "C"

import (
	"errors"
	"io"
	"unsafe"
)

type File C.FILE

const (
	EOF = C.EOF

	SEEK_CUR = C.SEEK_CUR
	SEEK_SET = C.SEEK_SET
	SEEK_END = C.SEEK_END
)

// OpenFile opens a C File Pointer
// mode can be specified with a file mode in the os package.
func OpenFile(filename, mode string) *File {
	return (*File)(C.fopen(C.CString(filename), C.CString(mode)))
}

func (f *File) Fileno() (int, error) {
	fileno := int(C.fileno((*C.FILE)(f)))
	if fileno == -1 {
		return fileno, errors.New(C.GoString(C.strerror(C.int(f.Error()))))
	}

	return fileno, nil
}

func (f *File) Read(buf []byte) (int, error) {
	n := int(C.fread(unsafe.Pointer(&buf[0]), C.size_t(1), C.size_t(len(buf)), (*C.FILE)(f)))

	if n > 0 {
		return n, nil
	}

	if f.Eof() != 0 {
		return 0, io.EOF
	}

	return 0, errors.New(C.GoString(C.strerror(C.int(f.Error()))))
}

func (f *File) Close() error {
	n := int(C.fclose((*C.FILE)(f)))

	if n != 0 {
		return io.EOF
	}

	return nil
}

func (f *File) Flush() {
	C.fflush((*C.FILE)(f))
}

func (f *File) Write(buf []byte) (int, error) {
	n := int(C.fwrite(unsafe.Pointer(&buf[0]), C.size_t(1), C.size_t(len(buf)), (*C.FILE)(f)))
	if n > 0 {
		return n, nil
	}

	if f.Eof() != 0 {
		return 0, io.EOF
	}

	return 0, errors.New(C.GoString(C.strerror(C.int(f.Error()))))
}

func (f *File) Seek(off int64, origin int) (int64, error) {
	n := int(C.fseek((*C.FILE)(f), C.long(off), C.int(origin)))
	if n == 0 {
		return f.Tell(), nil
	}

	return 0, errors.New(C.GoString(C.strerror(C.int(f.Error()))))
}

func (f *File) Eof() int {
	return int(C.feof((*C.FILE)(f)))
}

func (f *File) Error() int {
	return int(C.ferror((*C.FILE)(f)))
}

func (f *File) Tell() int64 {
	return int64(C.ftell((*C.FILE)(f)))
}
