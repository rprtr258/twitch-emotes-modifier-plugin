package repository

import (
	"os"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/repository/seventv"
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

func IsCached(emoteID string) bool {
	imageFilename := emoteID + ".webp"
	_, err := os.Stat(imageFilename)
	return err == nil
}

// TODO: load only if id
func Emote(emoteID string) (*webp.Animation, error) {
	if !IsCached(emoteID) {
		data, err := seventv.Repository{}.Download7tvEmote(emoteID, emoteID)
		if err != nil {
			return nil, err
		}

		if err := saveObject(data, emoteID); err != nil {
			return nil, err
		}
	}

	return loadObject(emoteID)
}
