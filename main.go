package main

import (
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tidbyt/go-libwebp/webp"
)

func loadEmote(filename string) (*webp.Animation, error) {
	data, err := os.ReadFile(filename)
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

func loadEmoteImage(emoteId string) (filename string, err error) {
	imageFilename := fmt.Sprintf("%s.%s", emoteId, "webp")

	if _, err := os.Stat(imageFilename); os.IsExist(err) {
		return imageFilename, nil
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

	return imageFilename, nil
}

type mergedTimestamp struct {
	timestamp int
	frames    []int
}

func unsafeMergeTimeSeries(first, second []int) []mergedTimestamp {
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

func over(firstFilename, secondFilename, outFilename string) error {
	firstImg, err := loadEmote(firstFilename)
	if err != nil {
		return err
	}

	secondImg, err := loadEmote(secondFilename)
	if err != nil {
		return err
	}

	mergedTimestamps := mergeTimeSeries(firstImg.Timestamp, secondImg.Timestamp)

	enc, err := webp.NewAnimationEncoder(firstImg.CanvasHeight, firstImg.CanvasWidth, 0, 0)
	if err != nil {
		return err
	}
	defer enc.Close()

	for i, ts := range mergedTimestamps {
		durationMillis := ts.timestamp
		if i > 0 {
			durationMillis -= mergedTimestamps[i-1].timestamp
		}

		firstFrame := firstImg.Image[ts.frames[0]]
		secondFrame := secondImg.Image[ts.frames[1]]

		buf := append([]uint8{}, firstFrame.Pix...)
		for i := 0; i < len(buf); i += 4 {
			// TODO: https://stackoverflow.com/questions/41093527/how-to-blend-two-rgb-unsigned-byte-colors-stored-as-unsigned-32bit-ints-fast
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
			return nil
		}
	}

	data, err := enc.Assemble()
	if err != nil {
		return err
	}

	if err := os.WriteFile(outFilename, data, 0666); err != nil {
		return err
	}

	return nil
}

func run() error {
	// 	emote := os.Args[1]
	// 	stack := []string{}
	// 	for _, token := range strings.Split(emote, ",") {
	// 		switch token {
	// 		case "^":
	// 			// TODO: assert stack size
	// 			fmt.Println("decoding", stack[len(stack)-2])
	// 			_, err := loadEmoteImage(stack[len(stack)-2])
	// 			if err != nil {
	// 				panic(err)
	// 			}
	// 			// secondEmote, err := loadEmoteImage(stack[len(stack)-2])
	// 			// if err != nil {
	// 			// 	panic(err)
	// 			// }
	// 			// fmt.Println(firstEmote, secondEmote)
	// 		default:
	// 			stack = append(stack, token)
	// 		}
	// 	}

	return over("peepoClap.webp", "snowTime.webp", "out.webp")
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}
