package modifiers

import (
	"image"
	"image/color"
	"math/rand"
	"time"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

type shiftedImage struct {
	img image.Image
	dx  int
	dy  int
}

func (im shiftedImage) ColorModel() color.Model {
	return im.img.ColorModel()
}

func (im shiftedImage) Bounds() image.Rectangle {
	return im.img.Bounds()
}

func (im shiftedImage) At(x, y int) color.Color {
	return im.img.At(x-im.dx, y-im.dy)
}

type Shake struct {
	// TODO: embed?
	In *webp.Animation
}

func (m Shake) Modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.In.CanvasWidth, m.In.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	for i, frame := range m.In.Image {
		newFrame := shiftedImage{
			img: frame,
			dx:  int(rand.Intn(m.In.CanvasWidth) - m.In.CanvasWidth/2),
			dy:  int(rand.Intn(m.In.CanvasHeight) - m.In.CanvasHeight/2),
		}

		if err := enc.AddFrame(newFrame, time.Duration(m.In.Timestamp[i])*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}
