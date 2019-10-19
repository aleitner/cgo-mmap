package cgommap_test

import (
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

	var buf []byte
	n, err = mmap.Read(buf)
	require.NoError(t, err)
	require.Equal(t, len(testText), len(buf))
	require.Equal(t, testText, buf)

	err = mmap.Close()
	require.NoError(t, err)
}