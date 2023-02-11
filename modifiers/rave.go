package modifiers

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

func rgb2hsv(r float64, g float64, b float64) (h float64, s float64, v float64) {
	// hue
	// ---------------------------------------------------------------------------------------------
	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)
	chroma := max - min

	if chroma == 0 {
		h = 0
	} else {
		var huePrime float64
		if r == max {
			huePrime = math.Mod(((g - b) / chroma), 6)
		} else if g == max {
			huePrime = ((b - r) / chroma) + 2
		} else if b == max {
			huePrime = ((r - g) / chroma) + 4
		}
		h = huePrime * 60
	}

	if h < 0 {
		h = 360 + h
	}

	// lightness
	// ---------------------------------------------------------------------------------------------
	if r == g && g == b {
		v = r
	} else {
		v = max
	}

	// saturation
	// ---------------------------------------------------------------------------------------------
	if v == 0 {
		s = 0
	} else {
		s = (chroma / v)
	}

	if math.IsNaN(s) {
		s = 0
	}

	return h, s, v
}

func hs2rgb(hueDegrees float64, saturation float64, lightOrVal float64) (uint16, uint16, uint16) {
	var r, g, b float64

	hueDegrees = math.Mod(hueDegrees, 360)

	if saturation == 0 {
		r = lightOrVal
		g = lightOrVal
		b = lightOrVal
	} else {
		chroma := lightOrVal * saturation

		hueSector := hueDegrees / 60

		intermediate := chroma * (1 - math.Abs(
			math.Mod(hueSector, 2)-1,
		))

		switch {
		case hueSector >= 0 && hueSector <= 1:
			r = chroma
			g = intermediate
			b = 0

		case hueSector > 1 && hueSector <= 2:
			r = intermediate
			g = chroma
			b = 0

		case hueSector > 2 && hueSector <= 3:
			r = 0
			g = chroma
			b = intermediate

		case hueSector > 3 && hueSector <= 4:
			r = 0
			g = intermediate
			b = chroma
		case hueSector > 4 && hueSector <= 5:
			r = intermediate
			g = 0
			b = chroma

		case hueSector > 5 && hueSector <= 6:
			r = chroma
			g = 0
			b = intermediate

		default:
			panic(fmt.Errorf("hue input %v yielded sector %v", hueDegrees, hueSector))
		}

		m := lightOrVal - chroma

		r += m
		g += m
		b += m
	}

	return uint16(r * 0xFFFF), uint16(g * 0xFFFF), uint16(b * 0xFFFF)
}

type hueImage struct {
	img image.Image
	hue float64
}

func (im hueImage) ColorModel() color.Model {
	return im.img.ColorModel()
}

func (im hueImage) Bounds() image.Rectangle {
	return im.img.Bounds()
}

func (im hueImage) At(x, y int) color.Color {
	r, g, b, a := im.img.At(x, y).RGBA()
	_, s, v := rgb2hsv(float64(r)/0xFFFF, float64(g)/0xFFFF, float64(b)/0xFFFF)
	r1, g1, b1 := hs2rgb(float64(im.hue), s, v)
	return color.RGBA64{r1, g1, b1, uint16(a)}
}

type Rave struct {
	// TODO: embed?
	In *webp.Animation
}

func (m Rave) Modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.In.CanvasWidth, m.In.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	totalTime := float64(m.In.Timestamp[len(m.In.Timestamp)-1])

	for i, frame := range m.In.Image {
		newFrame := hueImage{
			img: frame,
			hue: 360 * float64(m.In.Timestamp[i]) / totalTime,
		}

		if err := enc.AddFrame(newFrame, time.Duration(m.In.Timestamp[i])*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}
