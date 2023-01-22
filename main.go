package main

import (
	"fmt"
	"log"
	"os"
	"time"

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

func bench(message string) func() {
	start := time.Now()
	return func() {
		fmt.Println("done", message, "in", time.Since(start).String())
	}
}

func unaryTokenHandler(
	stack *stack,
	suffix string,
	construct func(*webp.Animation) modifiers.Modifier,
) error {
	arg := stack.pop()

	defer bench(fmt.Sprintf("%s %s", suffix, arg))()

	img, err := repository.Emote(arg)
	if err != nil {
		return err
	}

	m := construct(img)

	enc, err := m.Modify()
	if err != nil {
		return err
	}

	newEmoteID := arg + suffix

	if err := repository.SaveObject(enc, newEmoteID); err != nil {
		return err
	}

	stack.push(newEmoteID)

	return nil
}

func binaryTokenHandler(
	stack *stack,
	suffix string,
	construct func(a, b *webp.Animation) modifiers.Modifier,
) error {
	// TODO: assert stack size
	second := stack.pop()
	first := stack.pop()

	defer bench(fmt.Sprintf("%s %s %s", suffix, first, second))()

	firstImg, err := repository.Emote(first)
	if err != nil {
		return err
	}

	secondImg, err := repository.Emote(second)
	if err != nil {
		return err
	}

	m := construct(firstImg, secondImg)

	enc, err := m.Modify()
	if err != nil {
		return err
	}

	newEmote := fmt.Sprintf("%s,%s%s", first, second, suffix)

	// TODO: do not re-evaluate if already exists (cache)
	if err := repository.SaveObject(enc, newEmote); err != nil {
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
		rex.Common.Text(`>dup`),
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
		case ">dup":
			item := stack.pop()
			stack.push(item)
			stack.push(item)
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
