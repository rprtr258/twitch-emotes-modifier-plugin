package modifiers

import (
	"image"

	"github.com/gen2brain/webp"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal"
)

// TODO: maybe rename to mirror{x,y,t}
type ScaleX struct {
	// TODO: embed?
	In    *webp.WEBP
	Scale float64
}

func (m ScaleX) Modify() (*webp.WEBP, error) {
	first := internal.RGBA(m.In.Image[0])
	newWidth := int(float64(first.Rect.Dx()) * m.Scale)
	stride := newWidth * 4
	height := first.Rect.Dy()

	images := make([]image.Image, len(m.In.Image))
	for i, frame := range internal.ToRGBAs(m.In.Image) {
		buf := make([]uint8, stride*height)
		for j := 0; j < height; j++ {
			for i := 0; i < newWidth; i++ {
				for k := 0; k < 4; k++ {
					buf[j*stride+i*4+k] = frame.Pix[j*frame.Stride+int(float64(i)/m.Scale)*4+k]
				}
			}
		}

		images[i] = &image.RGBA{
			Pix:    buf,
			Stride: stride,
			Rect:   image.Rect(0, 0, newWidth, height),
		}
	}

	return &webp.WEBP{
		Image:     images,
		Delay:     m.In.Delay,
		LoopCount: m.In.LoopCount,
	}, nil
}

type ScaleY struct {
	In    *webp.WEBP
	Scale float64
}

func (m ScaleY) Modify() (*webp.WEBP, error) {
	first := internal.RGBA(m.In.Image[0])
	newHeight := int(float64(first.Rect.Dy()) * m.Scale)
	width := first.Rect.Dx()

	images := make([]image.Image, len(m.In.Image))
	for i, frame := range internal.ToRGBAs(m.In.Image) {
		buf := make([]uint8, frame.Stride*newHeight)
		for j := 0; j < newHeight; j++ {
			for i := 0; i < width; i++ {
				for k := 0; k < 4; k++ {
					buf[j*frame.Stride+i*4+k] = frame.Pix[int(float64(j)/m.Scale)*frame.Stride+i*4+k]
				}
			}
		}

		res := &image.RGBA{
			Pix:    buf,
			Stride: frame.Stride,
			Rect:   image.Rect(0, 0, width, newHeight),
		}

		images[i] = res
	}

	return &webp.WEBP{
		Image:     images,
		Delay:     m.In.Delay,
		LoopCount: m.In.LoopCount,
	}, nil
}

type ScaleT struct {
	In    *webp.WEBP
	Scale float64
}

func (m ScaleT) Modify() (_ *webp.WEBP, e error) {
	delays := make([]int, len(m.In.Delay))
	for i, d := range m.In.Delay {
		delays[i] = int(float64(d) * m.Scale)
	}

	return &webp.WEBP{
		Image:     m.In.Image,
		Delay:     delays,
		LoopCount: m.In.LoopCount,
	}, nil
}
