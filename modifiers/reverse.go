package modifiers

import (
	"image"

	"github.com/gen2brain/webp"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal"
)

// TODO: maybe rename to mirror{x,y,t}
type ReverseX struct {
	// TODO: embed?
	In *webp.WEBP
}

func (m ReverseX) Modify() (*webp.WEBP, error) {
	images := make([]image.Image, len(m.In.Image))
	for i, frame := range internal.ToRGBAs(m.In.Image) {
		buf := append([]uint8{}, frame.Pix...)
		for row := 0; row < frame.Rect.Dy(); row++ {
			stride := row * frame.Stride
			for i, j := 0, frame.Stride-4; i < j; i, j = i+4, j-4 {
				buf[stride+i+0], buf[stride+j+0] = buf[stride+j+0], buf[stride+i+0]
				buf[stride+i+1], buf[stride+j+1] = buf[stride+j+1], buf[stride+i+1]
				buf[stride+i+2], buf[stride+j+2] = buf[stride+j+2], buf[stride+i+2]
				buf[stride+i+3], buf[stride+j+3] = buf[stride+j+3], buf[stride+i+3]
			}
		}

		images[i] = &image.RGBA{
			Pix:    buf,
			Stride: frame.Stride,
			Rect:   frame.Rect,
		}
	}

	return &webp.WEBP{
		Image:     images,
		Delay:     m.In.Delay,
		LoopCount: m.In.LoopCount,
	}, nil
}

type ReverseY struct {
	In *webp.WEBP
}

func (m ReverseY) Modify() (*webp.WEBP, error) {
	images := make([]image.Image, len(m.In.Image))
	for i, frame := range internal.ToRGBAs(m.In.Image) {
		buf := append([]uint8{}, frame.Pix...)
		for i, j := 0, frame.Rect.Dy()-1; i < j; i, j = i+1, j-1 {
			strideI := i * frame.Stride
			strideJ := j * frame.Stride
			for k := 0; k < frame.Stride; k++ {
				buf[strideI+k], buf[strideJ+k] = buf[strideJ+k], buf[strideI+k]
			}
		}

		images[i] = &image.RGBA{
			Pix:    buf,
			Stride: frame.Stride,
			Rect:   frame.Rect,
		}
	}

	return &webp.WEBP{
		Image:     images,
		Delay:     m.In.Delay,
		LoopCount: m.In.LoopCount,
	}, nil
}

type ReverseT struct {
	In *webp.WEBP
}

func (m ReverseT) Modify() (*webp.WEBP, error) {
	timestamps := internal.ReverseTimestamps(m.In.Delay)

	images := make([]image.Image, len(m.In.Image))
	for i := range m.In.Image {
		images[i] = m.In.Image[len(m.In.Image)-i-1]
	}

	return &webp.WEBP{
		Image:     images,
		Delay:     timestamps,
		LoopCount: m.In.LoopCount,
	}, nil
}
