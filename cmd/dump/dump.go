package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tidbyt/go-libwebp/webp"
)

func loadEmote(filename string) (*webp.Animation, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	dec, err := webp.NewAnimationDecoder(data)
	if err != nil {
		return nil, err
	}
	defer dec.Close()

	anim, err := dec.Decode()
	if err != nil {
		return nil, err
	}

	return anim, nil
}

func durations(timestamps []int) []int {
	res := make([]int, len(timestamps))
	res[0] = timestamps[0]
	for i := 1; i < len(timestamps); i++ {
		res[i] = timestamps[i] - timestamps[i-1]
	}
	return res
}

func run() error {
	img, err := loadEmote(os.Args[1])
	if err != nil {
		return err
	}

	fmt.Println("Dimensions:", img.CanvasWidth, "*", img.CanvasHeight)
	fmt.Println("Frames:", img.FrameCount)
	fmt.Println("Durations:", durations(img.Timestamp))
	fmt.Println("Timestamps:", img.Timestamp)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}
