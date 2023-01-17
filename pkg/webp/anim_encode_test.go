package webp_test

import (
	"image"
	"testing"
	"time"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp/internal"
)

func TestEncodeAnimation(t *testing.T) {
	data := internal.ReadFile("cosmos.webp")
	aWebP, err := webp.DecodeRGBA(data, &webp.DecoderOptions{})
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}

	img := []image.Image{
		internal.ReadPNG("butterfly.png"),
		internal.ReadPNG("checkerboard.png"),
		internal.ReadPNG("yellow-rose-3.png"),
		aWebP,
	}

	width, height := 24, 24
	anim, err := webp.NewAnimationEncoder(width, height, 0, 0)
	if err != nil {
		t.Fatalf("initializing decoder: %v", err)
	}
	defer anim.Close()

	for i, im := range img {
		// all frames of an animation must have the same dimensions
		cropped := im.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(image.Rect(0, 0, width, height))

		if err := anim.AddFrame(cropped, 100*time.Millisecond); err != nil {
			t.Errorf("adding frame %d: %v", i, err)
		}
	}

	buf, err := anim.Assemble()
	if err != nil {
		t.Fatalf("assembling animation: %v", err)
	}
	if len(buf) == 0 {
		t.Errorf("assembled animation is empty")
	}
}
