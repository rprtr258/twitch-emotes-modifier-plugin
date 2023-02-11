package modifiers

import (
	"time"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

type SlideIn struct {
	// TODO: embed?
	In *webp.Animation
}

func (m SlideIn) Modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.In.CanvasWidth, m.In.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	totalTime := float64(m.In.Timestamp[len(m.In.Timestamp)-1])

	for i, frame := range m.In.Image {
		d := float64(m.In.Timestamp[i]) / totalTime
		newFrame := shiftedImage{
			img: frame,
			dx:  -int(float64(m.In.CanvasWidth) * (1 - d) * (1 - d)),
			dy:  0,
		}

		if err := enc.AddFrame(newFrame, time.Duration(m.In.Timestamp[i])*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}
