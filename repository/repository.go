package repository

import (
	"bytes"
	"os"

	"github.com/rprtr258/xerr"

	"github.com/gen2brain/webp"
)

type EmotesRepository struct{}

func objectErr(err error, message, objectID string) error {
	return xerr.New(
		xerr.Message(message),
		xerr.Errors{err},
		xerr.Fields{"objectID": objectID},
	)
}

func (EmotesRepository) LoadObject(objectID string) (*webp.WEBP, error) {
	data, err := os.Open(objectID + ".webp")
	if err != nil {
		return nil, objectErr(err, "failed loading object", objectID)
	}
	defer data.Close()

	anim, err := webp.DecodeAll(data)
	if err != nil {
		return nil, objectErr(err, "failed creating decoder", objectID)
	}

	return anim, nil
}

func (EmotesRepository) Save(data []byte, objectID string) error {
	if err := os.WriteFile(objectID+".webp", data, 0666); err != nil {
		return xerr.New(
			xerr.Errors{err},
			xerr.Fields{"objectID": objectID},
		)
	}

	return nil
}

func (r EmotesRepository) SaveObject(enc *webp.WEBP, objectID string) error {
	var data bytes.Buffer
	if err := webp.EncodeAll(&data, enc); err != nil {
		return err
	}

	if err := r.Save(data.Bytes(), objectID); err != nil {
		return err
	}

	return nil
}

func (EmotesRepository) IsCached(emoteID string) bool {
	imageFilename := emoteID + ".webp"
	_, err := os.Stat(imageFilename)
	return err == nil
}
