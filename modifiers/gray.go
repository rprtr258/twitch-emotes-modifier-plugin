package modifiers

import (
	"image"
	"image/color"
	"time"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

type grayscaleImage struct {
	img image.Image
}

func (im grayscaleImage) ColorModel() color.Model {
	return color.GrayModel
}

func (im grayscaleImage) Bounds() image.Rectangle {
	return im.img.Bounds()
}

func (im grayscaleImage) At(x, y int) color.Color {
	c := im.img.At(x, y)
	_, _, _, a := c.RGBA()
	if a == 0 {
		return c
	}
	return color.GrayModel.Convert(c)
}

type Gray struct {
	// TODO: embed?
	In *webp.Animation
}

func (m Gray) Modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.In.CanvasWidth, m.In.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	for i, frame := range m.In.Image {
		res := grayscaleImage{frame}

		if err := enc.AddFrame(res, time.Duration(m.In.Timestamp[i])*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}
