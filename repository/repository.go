package repository

import (
	"os"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

type EmotesRepository struct{}

func (EmotesRepository) LoadObject(objectID string) (*webp.Animation, error) {
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

func (EmotesRepository) Save(data []byte, objectID string) error {
	if err := os.WriteFile(objectID+".webp", data, 0666); err != nil {
		return err
	}

	return nil
}

func (r EmotesRepository) SaveObject(enc *webp.AnimationEncoder, objectID string) error {
	defer enc.Close()

	data, err := enc.Assemble()
	if err != nil {
		return err
	}

	if err := r.Save(data, objectID); err != nil {
		return err
	}

	return nil
}

func (EmotesRepository) IsCached(emoteID string) bool {
	imageFilename := emoteID + ".webp"
	_, err := os.Stat(imageFilename)
	return err == nil
}
