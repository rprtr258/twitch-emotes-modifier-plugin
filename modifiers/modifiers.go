package modifiers

import (
	"image"
	"image/color"

	"github.com/gen2brain/webp"
)

type Modifier interface {
	Modify() (*webp.WEBP, error)
}

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
