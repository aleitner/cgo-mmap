package cgommap_test

import (
	"bytes"
	"cgommap"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestMmap(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	file.Close()

	f := cgommap.OpenFile(file.Name(), cgommap.READWRITE_TRUNCATE)
	defer func() {
		err := f.Close()
		require.NoError(t, err)
	}()

	testText := []byte("Testing the memory mapped file")
	n, err := f.Write(testText)
	require.NoError(t, err)
	require.Equal(t, len(testText), n)

	f.Flush()

	fd, err := f.Fileno()
	require.NoError(t, err)

	mmap, err := cgommap.NewMmap(int64(len(testText)),0, cgommap.PROT_READ, cgommap.MAP_SHARED, fd)
	require.NoError(t, err)

	{ // Read Test
		buf := make([]byte, len(testText))
		n, err = mmap.Read(buf[0:])
		require.NoError(t, err)
		require.Equal(t, len(testText), len(buf))
		require.True(t, bytes.Equal(testText, buf))

		buf = make([]byte, len(testText))
		n, err = mmap.Read(buf[0:])
		require.Error(t, err)
	}


	{ // Seek test
		off, err := mmap.Seek(0, cgommap.SEEK_SET)
		require.NoError(t, err)
		require.Equal(t, int64(0), off)

		buf := make([]byte, len(testText))
		n, err = mmap.Read(buf[0:])
		require.NoError(t, err)
		require.Equal(t, len(testText), len(buf))
		require.True(t, bytes.Equal(testText, buf))
	}


	err = mmap.Close()
	require.NoError(t, err)
}