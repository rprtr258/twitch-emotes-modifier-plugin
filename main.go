package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hedhyw/rex/pkg/rex"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/modifiers"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/repository"
)

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
	emote, err := repository.DownloadEmote(arg)
	if err != nil {
		return err
	}

	// TODO: maybe use hash instead?
	newEmote := fmt.Sprintf("%s%s", emote, suffix)

	img, err := repository.LoadCachedEmote(emote)
	if err != nil {
		return err
	}

	m := construct(img)

	if err := modifiers.Run(m, newEmote); err != nil {
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
	firstEmote, err := repository.DownloadEmote(first)
	if err != nil {
		return err
	}

	secondEmote, err := repository.DownloadEmote(second)
	if err != nil {
		return err
	}

	firstImg, err := repository.LoadCachedEmote(firstEmote)
	if err != nil {
		return err
	}

	secondImg, err := repository.LoadCachedEmote(secondEmote)
	if err != nil {
		return err
	}

	m := construct(firstImg, secondImg)

	newEmote := fmt.Sprintf("%s,%s%s", first, second, suffix)

	// TODO: do not re-evaluate if already exists (cache)
	if err := modifiers.Run(m, newEmote); err != nil {
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
