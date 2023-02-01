// Package main is an example implementation of WebP decoder.
package main

import (
	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp/internal"
)

func main() {
	var err error

	// Read binary data
	data := internal.ReadFile("cosmos.webp")

	// Decode
	options := &webp.DecoderOptions{}
	img, err := webp.DecodeRGBA(data, options)
	if err != nil {
		panic(err)
	}

	internal.WritePNG(img, "encoded_cosmos.png")
}
