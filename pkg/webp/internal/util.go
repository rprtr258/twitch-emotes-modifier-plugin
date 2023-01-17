// Package util contains utility code for demosntration of go-libwebp.
package internal

import (
	"bufio"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

// GetExFilePath returns the path of specified example file.
func GetExFilePath(name string) string {
	path := filepath.Join("examples/images", name)
	must(os.Stat(path))
	return path
}

// GetOutFilePath returns the path of specified out file.
func GetOutFilePath(name string) string {
	path := "examples/out"
	must(os.Stat(path))
	return filepath.Join(path, name)
}

// OpenFile opens specified example file
func OpenFile(name string) (io io.Reader) {
	return must(os.Open(GetExFilePath(name)))
}

// ReadFile reads and returns data bytes of specified example file.
func ReadFile(name string) []byte {
	return must(os.ReadFile(GetExFilePath(name)))
}

// CreateFile opens specified example file
func CreateFile(name string) *os.File {
	return must(os.Create(GetOutFilePath(name)))
}

// WritePNG encodes and writes image into PNG file.
func WritePNG(img image.Image, name string) {
	f := must(os.Create(GetOutFilePath(name)))
	defer f.Close()
	b := bufio.NewWriter(f)
	defer b.Flush()

	if err := png.Encode(b, img); err != nil {
		panic(err)
	}
}

// ReadPNG reads and decodes png data into image.Image
func ReadPNG(name string) image.Image {
	return must(png.Decode(must(os.Open(GetExFilePath(name)))))
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
