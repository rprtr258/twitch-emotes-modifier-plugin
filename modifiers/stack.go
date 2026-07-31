package modifiers

import (
	"fmt"
	"image"

	"github.com/gen2brain/webp"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal"
)

type StackX struct {
	First, Second *webp.WEBP
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
func (m StackX) Modify() (*webp.WEBP, error) {
	first := internal.RGBA(m.First.Image[0])
	second := internal.RGBA(m.Second.Image[0])
	if first.Rect.Dy() != second.Rect.Dy() {
		return nil, fmt.Errorf("unequal heights on x-stack: %d and %d", first.Rect.Dy(), second.Rect.Dy())
	}

	mergedTimestamps := internal.MergeTimeSeries(m.First.Delay, m.Second.Delay)

	images := make([]image.Image, len(mergedTimestamps))
	delays := make([]int, len(mergedTimestamps))
	// TODO: cache same frames stacked
	for i, ts := range mergedTimestamps {
		images[i] = m.stack(
			internal.RGBA(m.First.Image[ts.Frames[0]]),
			internal.RGBA(m.Second.Image[ts.Frames[1]]),
		)
		delays[i] = ts.Timestamp
	}

	return &webp.WEBP{
		Image:     images,
		Delay:     delays,
		LoopCount: m.First.LoopCount * m.Second.LoopCount,
	}, nil
}

type StackY struct {
	First, Second *webp.WEBP
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

func (m StackY) Modify() (*webp.WEBP, error) {
	first := internal.RGBA(m.First.Image[0])
	second := internal.RGBA(m.Second.Image[0])
	if first.Rect.Dx() != second.Rect.Dx() {
		return nil, fmt.Errorf("unequal widths on y-stack: %d and %d", first.Rect.Dx(), second.Rect.Dx())
	}

	mergedTimestamps := internal.MergeTimeSeries(m.First.Delay, m.Second.Delay)

	// TODO: cache same frames stacked
	images := make([]image.Image, len(mergedTimestamps))
	delays := make([]int, len(mergedTimestamps))
	for i, ts := range mergedTimestamps {
		images[i] = m.stack(internal.RGBA(m.First.Image[ts.Frames[0]]), internal.RGBA(m.Second.Image[ts.Frames[1]]))
		delays[i] = ts.Timestamp
	}

	return &webp.WEBP{
		Image:     images,
		Delay:     delays,
		LoopCount: m.First.LoopCount * m.Second.LoopCount,
	}, nil
}

type StackT struct {
	First, Second *webp.WEBP
}

func (m StackT) Modify() (*webp.WEBP, error) {
	images := make([]image.Image, 0, len(m.First.Image)+len(m.Second.Image))
	delays := make([]int, 0, len(m.First.Image)+len(m.Second.Image))

	offset := m.First.Delay[len(m.First.Delay)-1]
	for i, frame := range m.First.Image {
		images = append(images, frame)
		delays = append(delays, m.First.Delay[i])
	}
	for i, frame := range m.Second.Image {
		images = append(images, frame)
		delays = append(delays, offset+m.Second.Delay[i])
	}

	return &webp.WEBP{
		Image:     images,
		Delay:     delays,
		LoopCount: m.First.LoopCount + m.Second.LoopCount,
	}, nil
}
