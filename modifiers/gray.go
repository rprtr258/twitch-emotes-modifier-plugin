package modifiers

import (
	"image"
	"image/color"

	"github.com/gen2brain/webp"
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
	In *webp.WEBP
}

func (m Gray) Modify() (*webp.WEBP, error) {
	images := make([]image.Image, len(m.In.Image))
	for i, frame := range m.In.Image {
		images[i] = grayscaleImage{frame}
	}

	return &webp.WEBP{
		Image:     images,
		Delay:     m.In.Delay,
		LoopCount: m.In.LoopCount,
	}, nil
}
