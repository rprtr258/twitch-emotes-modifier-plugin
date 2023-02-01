package modifiers

import (
	"image"
	"time"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

// TODO: maybe rename to stackz
type Over struct {
	First, Second *webp.Animation
}

func (m Over) Modify() (*webp.AnimationEncoder, error) {
	mergedTimestamps := internal.MergeTimeSeries(m.First.Timestamp, m.Second.Timestamp)

	enc, err := webp.NewAnimationEncoder(m.First.CanvasHeight, m.First.CanvasWidth, 0, 0)
	if err != nil {
		return nil, err
	}

	buf := make([]uint8, len(m.First.Image[0].Pix))
	for _, ts := range mergedTimestamps {
		firstFrame := m.First.Image[ts.Frames[0]]
		secondFrame := m.Second.Image[ts.Frames[1]]

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

		if err := enc.AddFrame(firstFrameCopy, time.Duration(ts.Timestamp)*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}
