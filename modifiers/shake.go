package modifiers

import (
	"image"
	"math/rand"

	"github.com/gen2brain/webp"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal"
)

type Shake struct {
	// TODO: embed?
	In *webp.WEBP
}

func (m Shake) Modify() (*webp.WEBP, error) {
	first := internal.RGBA(m.In.Image[0])
	width := first.Rect.Dx()
	height := first.Rect.Dy()

	images := make([]image.Image, len(m.In.Image))
	for i, frame := range m.In.Image {
		newFrame := shiftedImage{
			img: frame,
			dx:  int(rand.Intn(width) - width/2),
			dy:  int(rand.Intn(height) - height/2),
		}

		images[i] = newFrame
	}

	return &webp.WEBP{
		Image:     images,
		Delay:     m.In.Delay,
		LoopCount: m.In.LoopCount,
	}, nil
}
