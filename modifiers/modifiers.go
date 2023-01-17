package modifiers

import (
	"os"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

type Modifier interface {
	Modify() (*webp.AnimationEncoder, error)
}

func Run(m Modifier, outID string) error {
	enc, err := m.Modify()
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
