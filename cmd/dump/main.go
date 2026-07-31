package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gen2brain/webp"
)

func loadEmote(filename string) (*webp.WEBP, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	anim, err := webp.DecodeAll(f)
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

	fmt.Println("Dimensions:", img.Image[0].Bounds().Dx(), "*", img.Image[0].Bounds().Dy())
	fmt.Println("Frames:", len(img.Image))
	fmt.Println("Durations:", durations(img.Delay))
	fmt.Println("Timestamps:", img.Delay)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}
