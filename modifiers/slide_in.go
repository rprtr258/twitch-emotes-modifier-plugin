package modifiers

import (
	"image"

	"github.com/gen2brain/webp"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal"
)

type SlideIn struct {
	// TODO: embed?
	In *webp.WEBP
}

func (m SlideIn) Modify() (*webp.WEBP, error) {
	timestamps := internal.DelaysToTimestamps(m.In.Delay)
	totalTime := float64(timestamps[len(timestamps)-1])

	width := internal.RGBA(m.In.Image[0]).Rect.Dx()
	images := make([]image.Image, len(m.In.Image))
	for i, frame := range m.In.Image {
		d := float64(timestamps[i]) / totalTime
		newFrame := shiftedImage{
			img: frame,
			dx:  -int(float64(width) * (1 - d) * (1 - d)),
			dy:  0,
		}

		images[i] = newFrame
	}

	return &webp.WEBP{
		Image:     images,
		Delay:     m.In.Delay,
		LoopCount: m.In.LoopCount,
	}, nil
}
