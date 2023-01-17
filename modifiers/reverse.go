package modifiers

import (
	"image"
	"time"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

// TODO: maybe rename to mirror{x,y,t}
type ReverseX struct {
	// TODO: embed?
	In *webp.Animation
}

func (m ReverseX) Modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.In.CanvasWidth, m.In.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	for i := 0; i < m.In.FrameCount; i++ {
		durationMillis := m.In.Timestamp[i]
		if i > 0 {
			durationMillis -= m.In.Timestamp[i-1]
		}

		frame := m.In.Image[i]

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

		if err := enc.AddFrame(res, time.Duration(durationMillis)*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}

type ReverseY struct {
	In *webp.Animation
}

func (m ReverseY) Modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.In.CanvasWidth, m.In.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	for i := 0; i < m.In.FrameCount; i++ {
		durationMillis := m.In.Timestamp[i]
		if i > 0 {
			durationMillis -= m.In.Timestamp[i-1]
		}

		frame := m.In.Image[i]

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

		if err := enc.AddFrame(res, time.Duration(durationMillis)*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}

type ReverseT struct {
	In *webp.Animation
}

func (m ReverseT) Modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.In.CanvasWidth, m.In.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	for i := m.In.FrameCount - 1; i >= 0; i-- {
		durationMillis := m.In.Timestamp[i]
		if i > 0 {
			durationMillis -= m.In.Timestamp[i-1]
		}

		if err := enc.AddFrame(m.In.Image[i], time.Duration(durationMillis)*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}
