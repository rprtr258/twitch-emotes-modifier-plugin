package modifiers

import (
	"image"

	"github.com/gen2brain/webp"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal"
)

// TODO: maybe rename to stackz
type Over struct {
	First, Second *webp.WEBP
}

func (m Over) Modify() (*webp.WEBP, error) {
	mergedTimestamps := internal.MergeTimeSeries(m.First.Delay, m.Second.Delay)

	images := make([]image.Image, len(mergedTimestamps))
	delays := make([]int, len(mergedTimestamps))
	firstFrame0 := internal.RGBA(m.First.Image[0])
	buf := make([]uint8, len(firstFrame0.Pix))
	for i, ts := range mergedTimestamps {
		firstFrame := internal.RGBA(m.First.Image[ts.Frames[0]])
		secondFrame := internal.RGBA(m.Second.Image[ts.Frames[1]])

		buf := append(buf[:0], firstFrame.Pix...)
		for i := 0; i < len(buf); i += 4 {
			// TODO(OPTIMIZE): https://stackoverflow.com/questions/41093527/how-to-blend-two-rgb-unsigned-byte-colors-stored-as-unsigned-32bit-ints-fast
			alpha := int32(secondFrame.Pix[i+3])
			for j := 0; j < 3; j++ {
				a := int32(buf[i+j])
				b := int32(secondFrame.Pix[i+j])
				buf[i+j] = uint8((a*(255-alpha) + b*alpha) / 255)
			}
			if uint8(alpha) > buf[i+3] {
				buf[i+3] = uint8(alpha)
			}
		}
		firstFrameCopy := &image.RGBA{
			Pix:    buf,
			Stride: firstFrame.Stride,
			Rect:   firstFrame.Rect,
		}

		images[i] = firstFrameCopy
		delays[i] = ts.Timestamp
	}

	return &webp.WEBP{
		Image:     images,
		Delay:     delays,
		LoopCount: m.First.LoopCount * m.Second.LoopCount,
	}, nil
}
