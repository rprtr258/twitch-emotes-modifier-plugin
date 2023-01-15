package main

import (
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/tidbyt/go-libwebp/webp"
)

type modifier interface {
	modify() (*webp.AnimationEncoder, error)
}

func runModifier(m modifier, outID string) error {
	enc, err := m.modify()
	if err != nil {
		return err
	}
	defer enc.Close()

	data, err := enc.Assemble()
	if err != nil {
		return err
	}

	if err := os.WriteFile(outID+".webp", data, 0666); err != nil {
		return err
	}

	return nil
}

func loadEmote(id string) (*webp.Animation, error) {
	data, err := os.ReadFile(id + ".webp")
	if err != nil {
		return nil, err
	}

	dec, err := webp.NewAnimationDecoder(data)
	if err != nil {
		return nil, err
	}
	defer dec.Close()

	anim, err := dec.Decode()
	if err != nil {
		return nil, err
	}

	return anim, nil
}

func loadEmoteFilename(emoteId string) (filename string, err error) {
	imageFilename := fmt.Sprintf("%s.%s", emoteId, "webp")

	if _, err := os.Stat(imageFilename); err == nil {
		return emoteId, nil
	}

	emoteUrl := fmt.Sprintf("https://cdn.7tv.app/emote/%s/4x", emoteId)
	resp, err := http.Get(emoteUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// var extension string
	switch imageFormat := resp.Header.Get("Content-type"); imageFormat {
	case "image/webp":
		// extension = "webp"
	default:
		return "", fmt.Errorf("unknown image format: %s", imageFormat)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(imageFilename, data, 0666); err != nil {
		return "", err
	}

	return emoteId, nil
}

func unaryModifier(
	inID, outID string,
	construct func(*webp.Animation) modifier,
) error {
	img, err := loadEmote(inID)
	if err != nil {
		return err
	}

	m := construct(img)
	return runModifier(m, outID)
}

func binaryModifier(
	firstID, secondID, outID string,
	construct func(a, b *webp.Animation) modifier,
) error {
	firstImg, err := loadEmote(firstID)
	if err != nil {
		return err
	}

	secondImg, err := loadEmote(secondID)
	if err != nil {
		return err
	}

	m := construct(firstImg, secondImg)
	return runModifier(m, outID)
}

type mergedTimestamp struct {
	timestamp int
	frames    []int
}

func unsafeMergeTimeSeries(first, second []int) []mergedTimestamp {
	// if second is static
	if len(second) == 1 && second[0] == 0 {
		res := make([]mergedTimestamp, 0, len(first))
		for i, ts := range first {
			res = append(res, mergedTimestamp{
				timestamp: ts,
				frames:    []int{i, 0},
			})
		}
		return res
	}

	res := make([]mergedTimestamp, 0, len(first)+len(second))
	i, j := 0, 0
	secondOffset := 0
	for i < len(first) {
		var m mergedTimestamp
		switch {
		case first[i] < second[j]+secondOffset:
			m = mergedTimestamp{
				timestamp: first[i],
				frames:    []int{i, j},
			}
			i++
		case first[i] > second[j]+secondOffset:
			m = mergedTimestamp{
				timestamp: second[j] + secondOffset,
				frames:    []int{i, j},
			}
			j++
			if j == len(second) {
				j = 0
				secondOffset += second[len(second)-1]
			}
		case first[i] == second[j]+secondOffset:
			m = mergedTimestamp{
				timestamp: first[i],
				frames:    []int{i, j},
			}
			i++
			j++
			if j == len(second) {
				j = 0
				secondOffset += second[len(second)-1]
			}
		}
		res = append(res, m)
	}
	return res
}

func mergeTimeSeries(first, second []int) []mergedTimestamp {
	if len(first) == 0 || len(second) == 0 {
		panic("time series must not be empty")
	}

	if first[len(first)-1] < second[len(second)-1] {
		res := unsafeMergeTimeSeries(second, first)
		for i := range res {
			res[i].frames = []int{
				res[i].frames[1],
				res[i].frames[0],
			}
		}
		return res
	}

	return unsafeMergeTimeSeries(first, second)
}

type overModifier struct {
	first, second *webp.Animation
}

func (m overModifier) modify() (*webp.AnimationEncoder, error) {
	mergedTimestamps := mergeTimeSeries(m.first.Timestamp, m.second.Timestamp)

	enc, err := webp.NewAnimationEncoder(m.first.CanvasHeight, m.first.CanvasWidth, 0, 0)
	if err != nil {
		return nil, err
	}

	buf := make([]uint8, len(m.first.Image[0].Pix))
	for i, ts := range mergedTimestamps {
		durationMillis := ts.timestamp
		if i > 0 {
			durationMillis -= mergedTimestamps[i-1].timestamp
		}

		firstFrame := m.first.Image[ts.frames[0]]
		secondFrame := m.second.Image[ts.frames[1]]

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

		if err := enc.AddFrame(firstFrameCopy, time.Duration(durationMillis)*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}

type reverseModifier struct {
	in *webp.Animation
}

func (m reverseModifier) modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.in.CanvasWidth, m.in.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	for i := m.in.FrameCount - 1; i >= 0; i-- {
		durationMillis := m.in.Timestamp[i]
		if i > 0 {
			durationMillis -= m.in.Timestamp[i-1]
		}

		if err := enc.AddFrame(m.in.Image[i], time.Duration(durationMillis)*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}

type stack []string

func (s *stack) push(elem string) {
	*s = append(*s, elem)
}

func (s *stack) pop() string {
	if len(*s) == 0 {
		panic("can't pop from empty stack")
	}

	res := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return res
}

func run() error {
	tokenRE := regexp.MustCompile(`([-_A-Za-z():0-9]{2,99}|>over|>rev|,)`)
	stack := stack([]string{})
	for _, token := range tokenRE.FindAllString(os.Args[1], -1) {
		switch token {
		case ",":
		case ">rev":
			arg := stack.pop()
			emote, err := loadEmoteFilename(arg)
			if err != nil {
				return err
			}

			// TODO: maybe use hash instead?
			newEmote := fmt.Sprintf("%s>rev", emote)

			if err := unaryModifier(
				emote,
				newEmote,
				func(in *webp.Animation) modifier {
					return reverseModifier{
						in: in,
					}
				},
			); err != nil {
				return err
			}

			stack.push(newEmote)
		case ">over":
			second := stack.pop()
			first := stack.pop()

			// TODO: assert stack size
			// TODO: load only if id
			firstEmote, err := loadEmoteFilename(first)
			if err != nil {
				return err
			}

			secondEmote, err := loadEmoteFilename(second)
			if err != nil {
				return err
			}

			newEmote := fmt.Sprintf("%s,%s>over", first, second)
			// TODO: do not re-evaluate if already exists (cache)
			if err := binaryModifier(
				firstEmote,
				secondEmote,
				newEmote,
				func(a, b *webp.Animation) modifier {
					return overModifier{
						first:  a,
						second: b,
					}
				},
			); err != nil {
				return err
			}

			stack.push(newEmote)
		default:
			stack.push(token)
		}
	}

	if len(stack) != 1 {
		return fmt.Errorf("stack has more than single item (or none): %v", stack)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}
