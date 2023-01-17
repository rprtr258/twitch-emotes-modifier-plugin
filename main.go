package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/hedhyw/rex/pkg/rex"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/modifiers"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
)

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
	construct func(*webp.Animation) modifiers.Modifier,
) error {
	img, err := loadEmote(inID)
	if err != nil {
		return err
	}

	m := construct(img)
	return modifiers.Run(m, outID)
}

func binaryModifier(
	firstID, secondID, outID string,
	construct func(a, b *webp.Animation) modifiers.Modifier,
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
	return modifiers.Run(m, outID)
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
	construct func(*webp.Animation) modifiers.Modifier,
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
	construct func(a, b *webp.Animation) modifiers.Modifier,
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
	tokenRE := rex.New(rex.Group.Composite(
		rex.Common.Class( // match emote name
			rex.Chars.Alphanumeric(),
			rex.Chars.Runes("-_():"),
		).Repeat().Between(2, 99),
		rex.Common.Text(`,`),
		// match modifiers
		rex.Common.Text(`>over`),
		rex.Common.Text(`>revt`),
		rex.Common.Text(`>revx`),
		rex.Common.Text(`>revy`),
		rex.Common.Text(`>stackx`),
		rex.Common.Text(`>stacky`),
		rex.Common.Text(`>stackt`),
	)).MustCompile()
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
				func(in *webp.Animation) modifiers.Modifier {
					return modifiers.ReverseX{
						In: in,
					}
				},
			); err != nil {
				return err
			}
		case ">revy":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(in *webp.Animation) modifiers.Modifier {
					return modifiers.ReverseY{
						In: in,
					}
				},
			); err != nil {
				return err
			}
		case ">revt":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(in *webp.Animation) modifiers.Modifier {
					return modifiers.ReverseT{
						In: in,
					}
				},
			); err != nil {
				return err
			}
		case ">over":
			if err := binaryTokenHandler(
				&stack,
				token,
				func(a, b *webp.Animation) modifiers.Modifier {
					return modifiers.Over{
						First:  a,
						Second: b,
					}
				},
			); err != nil {
				return err
			}
		case ">stackx":
			if err := binaryTokenHandler(
				&stack,
				token,
				func(a, b *webp.Animation) modifiers.Modifier {
					return modifiers.StackX{
						First:  a,
						Second: b,
					}
				},
			); err != nil {
				return err
			}
		case ">stacky":
			if err := binaryTokenHandler(
				&stack,
				token,
				func(a, b *webp.Animation) modifiers.Modifier {
					return modifiers.StackY{
						First:  a,
						Second: b,
					}
				},
			); err != nil {
				return err
			}
		case ">stackt":
			if err := binaryTokenHandler(
				&stack,
				token,
				func(a, b *webp.Animation) modifiers.Modifier {
					return modifiers.StackT{
						First:  a,
						Second: b,
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
