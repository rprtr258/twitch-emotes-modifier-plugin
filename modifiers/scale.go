package modifiers

import (
	"image"
	"time"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

// TODO: maybe rename to mirror{x,y,t}
type ScaleX struct {
	// TODO: embed?
	In    *webp.Animation
	Scale float32
}

func (m ScaleX) Modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.In.CanvasWidth, m.In.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	for i, frame := range m.In.Image {
		buf := append([]uint8{}, frame.Pix...)
		for row := 0; row < m.In.CanvasHeight; row++ {
			stride := row * frame.Stride
			for i, j := 0, frame.Stride-4; i < j; i, j = i+4, j-4 {
				buf[stride+i+0], buf[stride+j+0] = buf[stride+j+0], buf[stride+i+0]
				buf[stride+i+1], buf[stride+j+1] = buf[stride+j+1], buf[stride+i+1]
				buf[stride+i+2], buf[stride+j+2] = buf[stride+j+2], buf[stride+i+2]
				buf[stride+i+3], buf[stride+j+3] = buf[stride+j+3], buf[stride+i+3]
			}
		}

		res := &image.RGBA{
			Pix:    buf,
			Stride: frame.Stride,
			Rect:   frame.Rect,
		}

		if err := enc.AddFrame(res, time.Duration(m.In.Timestamp[i])*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}

type ScaleY struct {
	In    *webp.Animation
	Scale float32
}

func (m ScaleY) Modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.In.CanvasWidth, m.In.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	for i, frame := range m.In.Image {
		buf := append([]uint8{}, frame.Pix...)
		for i, j := 0, m.In.CanvasHeight-1; i < j; i, j = i+1, j-1 {
			strideI := i * frame.Stride
			strideJ := j * frame.Stride
			for k := 0; k < frame.Stride; k++ {
				buf[strideI+k], buf[strideJ+k] = buf[strideJ+k], buf[strideI+k]
			}
		}

		res := &image.RGBA{
			Pix:    buf,
			Stride: frame.Stride,
			Rect:   frame.Rect,
		}

		if err := enc.AddFrame(res, time.Duration(m.In.Timestamp[i])*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}

type ScaleT struct {
	In    *webp.Animation
	Scale float32
}

func (m ScaleT) Modify() (_ *webp.AnimationEncoder, e error) {
	enc, err := webp.NewAnimationEncoder(m.In.CanvasWidth, m.In.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		if e != nil {
			enc.Close()
		}
	}()

	for i, frame := range m.In.Image {
		if err := enc.AddFrame(frame, time.Duration(float32(m.In.Timestamp[i])*m.Scale)*time.Millisecond); err != nil {
			return nil, err
		}
	}

	return enc, nil
}
