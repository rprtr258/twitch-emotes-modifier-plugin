// Package main is an example implementation of WebP encoder.
package main

import (
	"bufio"
	"image"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp/internal"
)

func main() {
	img := internal.ReadPNG("cosmos.png")

	// Create file and buffered writer
	io := internal.CreateFile("encoded_cosmos.webp")
	w := bufio.NewWriter(io)
	defer func() {
		w.Flush()
		io.Close()
	}()

	config, err := webp.ConfigPreset(webp.PresetDefault, 90)
	if err != nil {
		panic(err)
	}

	// Encode into WebP
	if err := webp.EncodeRGBA(w, img.(*image.RGBA), config); err != nil {
		panic(err)
	}
}
