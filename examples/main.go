package main

import (
	"cgommap"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	exampleText := []byte("Example text in a memory mapped file!")
	_, err = file.Write(exampleText)
	if err != nil {
		panic(err)
	}


	mmap, err := cgommap.New(int64(len(exampleText)),0, cgommap.PROT_READWRITE, cgommap.MAP_SHARED, file.Fd())
	if err != nil {
		panic(err)
	}

	defer mmap.Close()

	buf := make([]byte, mmap.Size())
	_, err = mmap.Read(buf[0:])

	fmt.Println(string(buf))
}