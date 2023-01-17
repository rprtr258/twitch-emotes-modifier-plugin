// Package util contains utility code for demosntration of go-libwebp.
package internal

import (
	"bufio"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// GetExFilePath returns the path of specified example file.
func GetExFilePath(name string) string {
	path := filepath.Join("../examples/images", name)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	panic(fmt.Errorf("%v does not exist", path))
}

// GetOutFilePath returns the path of specified out file.
func GetOutFilePath(name string) string {
	path := "../examples/out"
	if _, err := os.Stat(path); err == nil {
		return filepath.Join(path, name)
	}
	panic(fmt.Errorf("out directory does not exist"))
}

// OpenFile opens specified example file
func OpenFile(name string) (io io.Reader) {
	io, err := os.Open(GetExFilePath(name))
	if err != nil {
		panic(err)
	}
	return
}

// ReadFile reads and returns data bytes of specified example file.
func ReadFile(name string) (data []byte) {
	data, err := ioutil.ReadFile(GetExFilePath(name))
	if err != nil {
		panic(err)
	}
	return
}

// CreateFile opens specified example file
func CreateFile(name string) (f *os.File) {
	f, err := os.Create(GetOutFilePath(name))
	if err != nil {
		panic(err)
	}
	return
}

// WritePNG encodes and writes image into PNG file.
func WritePNG(img image.Image, name string) {
	f, err := os.Create(GetOutFilePath(name))
	if err != nil {
		panic(err)
	}
	b := bufio.NewWriter(f)
	defer func() {
		b.Flush()
		f.Close()
	}()

	if err := png.Encode(b, img); err != nil {
		panic(err)
	}
	return
}

// ReadPNG reads and decodes png data into image.Image
func ReadPNG(name string) (img image.Image) {
	io, err := os.Open(GetExFilePath(name))
	if err != nil {
		panic(err)
	}
	img, err = png.Decode(io)
	if err != nil {
		panic(err)
	}
	return
}
