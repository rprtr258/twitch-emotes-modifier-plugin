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

func run() error {
	img, err := loadEmote(os.Args[1])
	if err != nil {
		return err
	}

	fmt.Println("Dimensions:", img.CanvasWidth, "*", img.CanvasHeight)
	fmt.Println("Frames:", img.FrameCount)
	fmt.Println("Timestamp:", img.Timestamp)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}
