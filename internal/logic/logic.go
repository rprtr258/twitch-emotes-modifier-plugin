package logic

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hedhyw/rex/pkg/rex"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/modifiers"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/repository"
)

func hash(request string) string {
	h := sha1.New()
	h.Write([]byte(request))
	res := hex.EncodeToString(h.Sum(nil))
	log.Println(res, "IS", request)
	return res
}

type stack struct {
	elems []string
	err   error
}

func newStack() stack {
	return stack{
		elems: nil,
		err:   nil,
	}
}

func (s stack) Err() error {
	return s.err
}

func (s stack) len() int {
	if s.err != nil {
		return 0
	}

	return len(s.elems)
}

func (s *stack) push(elem string) {
	if s.err != nil {
		return
	}

	s.elems = append(s.elems, elem)
}

func (s *stack) popNum() float32 {
	if s.err != nil {
		return 0
	}

	if len(s.elems) == 0 {
		s.err = errors.New("tried to pop number from empty stack")
		return 0
	}

	poped := s.elems[len(s.elems)-1]
	s.elems = s.elems[:len(s.elems)-1]

	res, err := strconv.ParseFloat(poped, 32)
	if err != nil {
		s.err = fmt.Errorf("failed parsing number: %w", err)
		return 0
	}

	return float32(res)
}

func (s *stack) pop() string {
	if s.err != nil {
		return ""
	}

	if len(s.elems) == 0 {
		s.err = errors.New("tried to pop from empty stack")
		return ""
	}

	res := s.elems[len(s.elems)-1]
	s.elems = s.elems[:len(s.elems)-1]
	return res
}

func bench(message string) func() {
	start := time.Now()
	return func() {
		fmt.Println("done", message, "in", time.Since(start).String())
	}
}

func linearTokenHandler(
	stack *stack,
	suffix string,
	construct func(float32, *webp.Animation) modifiers.Modifier,
) error {
	// TODO: check stack errors after all pops
	arg := stack.pop()
	coeff := stack.popNum()

	defer bench(fmt.Sprintf("%s %s", suffix, arg))()

	newEmote := hash(arg + suffix + strconv.FormatFloat(float64(coeff), 'f', -1, 32))
	if !repository.IsCached(newEmote) {
		img, err := repository.Emote(arg)
		if err != nil {
			return err
		}

		m := construct(coeff, img)

		enc, err := m.Modify()
		if err != nil {
			return err
		}

		if err := repository.SaveObject(enc, newEmote); err != nil {
			return err
		}
	}

	stack.push(newEmote)

	return nil
}

func unaryTokenHandler(
	stack *stack,
	suffix string,
	construct func(*webp.Animation) modifiers.Modifier,
) error {
	arg := stack.pop()

	defer bench(fmt.Sprintf("%s %s", suffix, arg))()

	newEmote := hash(arg + suffix)
	if !repository.IsCached(newEmote) {
		img, err := repository.Emote(arg)
		if err != nil {
			return err
		}

		m := construct(img)

		enc, err := m.Modify()
		if err != nil {
			return err
		}

		if err := repository.SaveObject(enc, newEmote); err != nil {
			return err
		}
	}

	stack.push(newEmote)

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

	newEmote := hash(fmt.Sprintf("%s,%s%s", first, second, suffix))
	if !repository.IsCached(newEmote) {
		m := construct(firstImg, secondImg)

		enc, err := m.Modify()
		if err != nil {
			return err
		}

		// TODO: do not re-evaluate if already exists (cache)
		if err := repository.SaveObject(enc, newEmote); err != nil {
			return err
		}
	}

	stack.push(newEmote)

	return nil
}

func ProcessQuery(query string) (string, error) {
	tokenRE := rex.New(rex.Group.Composite(
		// match emote name
		rex.Common.Class(
			rex.Chars.Alphanumeric(),
			rex.Chars.Runes("-_():"),
		).Repeat().Between(2, 99),
		rex.Common.Raw(`[0-9]+\.?[0-9]*`),
		// ids separator
		rex.Common.Text(`,`),
		// match modifiers
		rex.Common.Text(`>over`),
		rex.Common.Text(`>revx`),
		rex.Common.Text(`>revy`),
		rex.Common.Text(`>revt`),
		rex.Common.Text(`>stackx`),
		rex.Common.Text(`>stacky`),
		rex.Common.Text(`>stackt`),
		rex.Common.Text(`>scalex`),
		rex.Common.Text(`>scaley`),
		rex.Common.Text(`>scalet`),
		rex.Common.Text(`>iscalex`),
		rex.Common.Text(`>iscaley`),
		rex.Common.Text(`>iscalet`),
		rex.Common.Text(`>dscalex`),
		rex.Common.Text(`>dscaley`),
		rex.Common.Text(`>dscalet`),
		rex.Common.Text(`>dup`),
		rex.Common.Text(`>swap`),
		rex.Common.Text(`>gray`),
		rex.Common.Text(`>shake`),
		rex.Common.Text(`>slide_in`),
		rex.Common.Text(`>rave`),
	)).MustCompile()
	stack := newStack()
	// TODO: assert all characters are used in tokenizing
	for _, token := range tokenRE.FindAllString(query, -1) {
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
				return "", err
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
				return "", err
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
				return "", err
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
				return "", err
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
				return "", err
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
				return "", err
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
				return "", err
			}
		case ">scalex":
			if err := linearTokenHandler(
				&stack,
				token,
				func(coeff float32, x *webp.Animation) modifiers.Modifier {
					return modifiers.ScaleX{
						In:    x,
						Scale: coeff,
					}
				},
			); err != nil {
				return "", err
			}
		case ">scaley":
			if err := linearTokenHandler(
				&stack,
				token,
				func(coeff float32, x *webp.Animation) modifiers.Modifier {
					return modifiers.ScaleY{
						In:    x,
						Scale: coeff,
					}
				},
			); err != nil {
				return "", err
			}
		case ">scalet":
			if err := linearTokenHandler(
				&stack,
				token,
				func(coeff float32, x *webp.Animation) modifiers.Modifier {
					return modifiers.ScaleT{
						In:    x,
						Scale: coeff,
					}
				},
			); err != nil {
				return "", err
			}
		case ">iscalex":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(x *webp.Animation) modifiers.Modifier {
					return modifiers.ScaleX{
						In:    x,
						Scale: 2.0,
					}
				},
			); err != nil {
				return "", err
			}
		case ">iscaley":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(x *webp.Animation) modifiers.Modifier {
					return modifiers.ScaleY{
						In:    x,
						Scale: 2.0,
					}
				},
			); err != nil {
				return "", err
			}
		case ">iscalet":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(x *webp.Animation) modifiers.Modifier {
					return modifiers.ScaleT{
						In:    x,
						Scale: 2.0,
					}
				},
			); err != nil {
				return "", err
			}
		case ">dscalex":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(x *webp.Animation) modifiers.Modifier {
					return modifiers.ScaleX{
						In:    x,
						Scale: 0.5,
					}
				},
			); err != nil {
				return "", err
			}
		case ">dscaley":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(x *webp.Animation) modifiers.Modifier {
					return modifiers.ScaleY{
						In:    x,
						Scale: 0.5,
					}
				},
			); err != nil {
				return "", err
			}
		case ">dscalet":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(x *webp.Animation) modifiers.Modifier {
					return modifiers.ScaleT{
						In:    x,
						Scale: 0.5,
					}
				},
			); err != nil {
				return "", err
			}
		case ">dup":
			item := stack.pop()
			stack.push(item)
			stack.push(item)
		case ">swap":
			first := stack.pop()
			second := stack.pop()
			stack.push(first)
			stack.push(second)
		case ">gray":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(x *webp.Animation) modifiers.Modifier {
					return modifiers.Gray{
						In: x,
					}
				},
			); err != nil {
				return "", err
			}
		case ">shake":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(x *webp.Animation) modifiers.Modifier {
					return modifiers.Shake{
						In: x,
					}
				},
			); err != nil {
				return "", err
			}
		case ">slide_in":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(x *webp.Animation) modifiers.Modifier {
					return modifiers.SlideIn{
						In: x,
					}
				},
			); err != nil {
				return "", err
			}
		case ">rave":
			if err := unaryTokenHandler(
				&stack,
				token,
				func(x *webp.Animation) modifiers.Modifier {
					return modifiers.Rave{
						In: x,
					}
				},
			); err != nil {
				return "", err
			}
		default:
			stack.push(token)
		}
	}

	if err := stack.Err(); err != nil {
		return "", err
	}

	if stack.len() != 1 {
		return "", fmt.Errorf("stack has more than single item (or none): %v", stack)
	}

	return stack.pop(), nil
}
