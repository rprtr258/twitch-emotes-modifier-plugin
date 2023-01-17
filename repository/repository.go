package repository

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

func loadObject(objectID string) (*webp.Animation, error) {
	data, err := os.ReadFile(objectID + ".webp")
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

func saveObject(data []byte, objectID string) error {
	if err := os.WriteFile(objectID+".webp", data, 0666); err != nil {
		return err
	}

	return nil
}

func SaveObject(enc *webp.AnimationEncoder, objectID string) error {
	defer enc.Close()

	data, err := enc.Assemble()
	if err != nil {
		return err
	}

	if err := saveObject(data, objectID); err != nil {
		return err
	}

	return nil
}

func getStoredID(emoteID string) (string, bool) {
	imageFilename := emoteID + ".webp"

	if _, err := os.Stat(imageFilename); err == nil {
		return imageFilename, true
	}

	return "", false
}

func download7tvEmote(emoteID, outObjectID string) ([]byte, error) {
	emoteUrl := fmt.Sprintf("https://cdn.7tv.app/emote/%s/4x", emoteID)
	resp, err := http.Get(emoteUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// var extension string
	switch imageFormat := resp.Header.Get("Content-type"); imageFormat {
	case "image/webp":
		// extension = "webp"
	default:
		return nil, fmt.Errorf("unknown image format: %s", imageFormat)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// TODO: load only if id
func Emote(emoteID string) (*webp.Animation, error) {
	imageFilename, ok := getStoredID(emoteID)
	if !ok {
		data, err := download7tvEmote(emoteID, imageFilename)
		if err != nil {
			return nil, err
		}

		if err := saveObject(data, emoteID); err != nil {
			return nil, err
		}
	}

	return loadObject(imageFilename)
}
