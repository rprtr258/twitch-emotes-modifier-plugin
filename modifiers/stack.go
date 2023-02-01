package modifiers

import (
	"fmt"
	"image"
	"time"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

type StackX struct {
	First, Second *webp.Animation
}

func (m StackX) stack(a, b *image.RGBA) *image.RGBA {
	buf := make([]uint8, 0, len(a.Pix)+len(b.Pix))
	for i := 0; i < a.Rect.Dy(); i++ {
		buf = append(buf, a.Pix[i*a.Stride:][:a.Stride]...)
		buf = append(buf, b.Pix[i*b.Stride:][:b.Stride]...)
	}

	return &image.RGBA{
		Pix:    buf,
		Stride: a.Stride + b.Stride,
		Rect:   image.Rect(0, 0, a.Rect.Dx()+b.Rect.Dx(), a.Rect.Dy()),
	}
}

// TODO: fix animation slowdown for some reason for >dup>revt>stackx and >dup>revt>stacky
func (m StackX) Modify() (*webp.AnimationEncoder, error) {
	if m.First.CanvasHeight != m.Second.CanvasHeight {
		return nil, fmt.Errorf("unequal heights on x-stack: %d and %d", m.First.CanvasHeight, m.Second.CanvasHeight)
	}

	enc, err := webp.NewAnimationEncoder(m.First.CanvasWidth+m.Second.CanvasWidth, m.First.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	mergedTimestamps := internal.MergeTimeSeries(m.First.Timestamp, m.Second.Timestamp)

	// TODO: cache same frames stacked
	for _, ts := range mergedTimestamps {
		frame := m.stack(
			m.First.Image[ts.Frames[0]],
			m.Second.Image[ts.Frames[1]],
		)
		if err := enc.AddFrame(frame, time.Duration(ts.Timestamp)*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}

type StackY struct {
	First, Second *webp.Animation
}

func (m StackY) stack(a, b *image.RGBA) *image.RGBA {
	buf := make([]uint8, 0, len(a.Pix)+len(b.Pix))
	buf = append(buf, a.Pix...)
	buf = append(buf, b.Pix...)

	return &image.RGBA{
		Pix:    buf,
		Stride: a.Stride,
		Rect:   image.Rect(0, 0, a.Rect.Dx(), a.Rect.Dy()+b.Rect.Dy()),
	}
}

func (m StackY) Modify() (*webp.AnimationEncoder, error) {
	if m.First.CanvasWidth != m.Second.CanvasWidth {
		return nil, fmt.Errorf("unequal widths on y-stack: %d and %d", m.First.CanvasWidth, m.Second.CanvasWidth)
	}

	enc, err := webp.NewAnimationEncoder(m.First.CanvasWidth, m.First.CanvasHeight+m.Second.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	mergedTimestamps := internal.MergeTimeSeries(m.First.Timestamp, m.Second.Timestamp)

	// TODO: cache same frames stacked
	for _, ts := range mergedTimestamps {
		frame := m.stack(m.First.Image[ts.Frames[0]], m.Second.Image[ts.Frames[1]])
		if err := enc.AddFrame(frame, time.Duration(ts.Timestamp)*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}

type StackT struct {
	First, Second *webp.Animation
}

func (m StackT) append(enc *webp.AnimationEncoder, img *webp.Animation, offset int) error {
	for i, frame := range img.Image {
		if err := enc.AddFrame(frame, time.Duration(img.Timestamp[i]+offset)*time.Millisecond); err != nil {
			enc.Close()
			return err
		}
	}

	return nil
}

func (m StackT) Modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.First.CanvasWidth, m.First.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	if err := m.append(enc, m.First, 0); err != nil {
		return nil, err
	}

	if err := m.append(enc, m.Second, m.First.Timestamp[m.First.FrameCount-1]); err != nil {
		return nil, err
	}

	return enc, nil
}
