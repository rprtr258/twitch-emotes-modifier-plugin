package modifiers

import (
	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

type Modifier interface {
	Modify() (*webp.AnimationEncoder, error)
}
