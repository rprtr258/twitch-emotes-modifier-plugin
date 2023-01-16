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

type stackXModifier struct {
	first, second *webp.Animation
}

func (m stackXModifier) stack(a, b *image.RGBA) *image.RGBA {
	buf := make([]uint8, 0, len(a.Pix)+len(b.Pix))
	for i := 0; i < a.Rect.Dy(); i++ {
		buf = append(buf, a.Pix[i*a.Stride:(i+1)*a.Stride]...)
		buf = append(buf, b.Pix[i*b.Stride:(i+1)*b.Stride]...)
	}

	return &image.RGBA{
		Pix:    buf,
		Stride: a.Stride + b.Stride,
		Rect:   image.Rect(0, 0, a.Rect.Dx()+b.Rect.Dx(), a.Rect.Dy()),
	}
}

func (m stackXModifier) modify() (*webp.AnimationEncoder, error) {
	if m.first.CanvasHeight != m.second.CanvasHeight {
		return nil, fmt.Errorf("unequal heights on x-stack: %d and %d", m.first.CanvasHeight, m.second.CanvasHeight)
	}

	enc, err := webp.NewAnimationEncoder(m.first.CanvasWidth+m.second.CanvasWidth, m.first.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	mergedTimestamps := mergeTimeSeries(m.first.Timestamp, m.second.Timestamp)

	// TODO: cache same frames stacked
	for i, ts := range mergedTimestamps {
		durationMillis := ts.timestamp
		if i > 0 {
			durationMillis -= mergedTimestamps[i-1].timestamp
		}

		frame := m.stack(m.first.Image[ts.frames[0]], m.second.Image[ts.frames[1]])
		if err := enc.AddFrame(frame, time.Duration(durationMillis)*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}

type stackYModifier struct {
	first, second *webp.Animation
}

func (m stackYModifier) stack(a, b *image.RGBA) *image.RGBA {
	buf := make([]uint8, 0, len(a.Pix)+len(b.Pix))
	buf = append(buf, a.Pix...)
	buf = append(buf, b.Pix...)

	return &image.RGBA{
		Pix:    buf,
		Stride: a.Stride,
		Rect:   image.Rect(0, 0, a.Rect.Dx(), a.Rect.Dy()+b.Rect.Dy()),
	}
}

func (m stackYModifier) modify() (*webp.AnimationEncoder, error) {
	if m.first.CanvasWidth != m.second.CanvasWidth {
		return nil, fmt.Errorf("unequal widths on y-stack: %d and %d", m.first.CanvasWidth, m.second.CanvasWidth)
	}

	enc, err := webp.NewAnimationEncoder(m.first.CanvasWidth, m.first.CanvasHeight+m.second.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	mergedTimestamps := mergeTimeSeries(m.first.Timestamp, m.second.Timestamp)

	// TODO: cache same frames stacked
	for i, ts := range mergedTimestamps {
		durationMillis := ts.timestamp
		if i > 0 {
			durationMillis -= mergedTimestamps[i-1].timestamp
		}

		frame := m.stack(m.first.Image[ts.frames[0]], m.second.Image[ts.frames[1]])
		if err := enc.AddFrame(frame, time.Duration(durationMillis)*time.Millisecond); err != nil {
			enc.Close()
			return nil, err
		}
	}

	return enc, nil
}

type stackTModifier struct {
	first, second *webp.Animation
}

func (m stackTModifier) append(enc *webp.AnimationEncoder, img *webp.Animation) error {
	for i, frame := range img.Image {
		durationMillis := img.Timestamp[i]
		if i > 0 {
			durationMillis -= img.Timestamp[i-1]
		}

		if err := enc.AddFrame(frame, time.Duration(durationMillis)*time.Millisecond); err != nil {
			enc.Close()
			return err
		}
	}

	return nil
}

func (m stackTModifier) modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.first.CanvasWidth, m.first.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	if err := m.append(enc, m.first); err != nil {
		return nil, err
	}

	if err := m.append(enc, m.second); err != nil {
		return nil, err
	}

	return enc, nil
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

type reverseTModifier struct {
	in *webp.Animation
}

func (m reverseTModifier) modify() (*webp.AnimationEncoder, error) {
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

type reverseXModifier struct {
	// TODO: embed?
	in *webp.Animation
}

func (m reverseXModifier) modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.in.CanvasWidth, m.in.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	for i := 0; i < m.in.FrameCount; i++ {
		durationMillis := m.in.Timestamp[i]
		if i > 0 {
			durationMillis -= m.in.Timestamp[i-1]
		}

		frame := m.in.Image[i]

		buf := append([]uint8{}, frame.Pix...)
		for row := 0; row < m.in.CanvasHeight; row++ {
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

type reverseYModifier struct {
	in *webp.Animation
}

func (m reverseYModifier) modify() (*webp.AnimationEncoder, error) {
	enc, err := webp.NewAnimationEncoder(m.in.CanvasWidth, m.in.CanvasHeight, 0, 0)
	if err != nil {
		return nil, err
	}

	for i := 0; i < m.in.FrameCount; i++ {
		durationMillis := m.in.Timestamp[i]
		if i > 0 {
			durationMillis -= m.in.Timestamp[i-1]
		}

		frame := m.in.Image[i]

		buf := append([]uint8{}, frame.Pix...)
		for i, j := 0, m.in.CanvasHeight-1; i < j; i, j = i+1, j-1 {
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

func unaryTokenHandler(
	stack *stack,
	suffix string,
	construct func(*webp.Animation) modifier,
) error {
	arg := stack.pop()
	emote, err := loadEmoteFilename(arg)
	if err != nil {
		return err
	}

	// TODO: maybe use hash instead?
	newEmote := fmt.Sprintf("%s%s", emote, suffix)

	if err := unaryModifier(emote, newEmote, construct); err != nil {
		return err
	}

	stack.push(newEmote)

	return nil
}

func binaryTokenHandler(
	stack *stack,
	suffix string,
	construct func(a, b *webp.Animation) modifier,
) error {
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

	newEmote := fmt.Sprintf("%s,%s%s", first, second, suffix)
	// TODO: do not re-evaluate if already exists (cache)
	if err := binaryModifier(
		firstEmote, secondEmote, newEmote,
		construct,
	); err != nil {
		return err
	}

	stack.push(newEmote)

	return nil
}

func run() error {
	tokenRE := regexp.MustCompile(`([-_A-Za-z():0-9]{2,99}|>over|>revt|>revx|>revy|>stackx|>stacky|>stackt|,)`)
	stack := stack([]string{})
	// TODO: assert all characters are used in tokenizing
	for _, token := range tokenRE.FindAllString(os.Args[1], -1) {
		fmt.Println("TOKEN", token)
		switch token {
		case ",":
		case ">revx":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(in *webp.Animation) modifier {
					return reverseXModifier{
						in: in,
					}
				},
			); err != nil {
				return err
			}
		case ">revy":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(in *webp.Animation) modifier {
					return reverseYModifier{
						in: in,
					}
				},
			); err != nil {
				return err
			}
		case ">revt":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(in *webp.Animation) modifier {
					return reverseTModifier{
						in: in,
					}
				},
			); err != nil {
				return err
			}
		case ">over":
			if err := binaryTokenHandler(
				&stack,
				token,
				func(a, b *webp.Animation) modifier {
					return overModifier{
						first:  a,
						second: b,
					}
				},
			); err != nil {
				return err
			}
		case ">stackx":
			if err := binaryTokenHandler(
				&stack,
				token,
				func(a, b *webp.Animation) modifier {
					return stackXModifier{
						first:  a,
						second: b,
					}
				},
			); err != nil {
				return err
			}
		case ">stacky":
			if err := binaryTokenHandler(
				&stack,
				token,
				func(a, b *webp.Animation) modifier {
					return stackYModifier{
						first:  a,
						second: b,
					}
				},
			); err != nil {
				return err
			}
		case ">stackt":
			if err := binaryTokenHandler(
				&stack,
				token,
				func(a, b *webp.Animation) modifier {
					return stackTModifier{
						first:  a,
						second: b,
					}
				},
			); err != nil {
				return err
			}
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
