package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tidbyt/go-libwebp/webp"
)

// // TODO: caching
// func loadEmoteImage(emoteId string) (image.Image, error) {
// 	// emoteUrl := fmt.Sprintf("https://cdn.7tv.app/emote/%s/4x", emoteId)
// 	// resp, err := http.Get(emoteUrl)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// defer resp.Body.Close()
// 	// imageFormat := resp.Header.Get("Content-type")
// 	// var extension string
// 	// switch imageFormat {
// 	// case "image/webp":
// 	// 	extension = "webp"
// 	// default:
// 	// 	return nil, fmt.Errorf("unknown image format: %s", imageFormat)
// 	// }
// 	imageFilename := fmt.Sprintf("%s.%s", emoteId, "webp" /*extension*/)
// 	// f, err := os.Create(imageFilename)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// defer f.Close()
// 	// io.Copy(f, resp.Body)
// 	readF, err := os.Open(imageFilename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer readF.Close()
// 	return DecodeWebp(readF)
// }

// func main() {
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
// }

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

func run() error {
	peepoClap, err := loadEmote("peepoClap.webp")
	if err != nil {
		return err
	}

	snowTime, err := loadEmote("snowTime.webp")
	if err != nil {
		return err
	}

	fmt.Println(peepoClap.Timestamp, snowTime.Timestamp)

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}
