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

	"github.com/rprtr258/xerr"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/modifiers"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/pkg/webp"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/repository"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/repository/seventv"
)

var (
	modifierRE = rex.Group.Composite(
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
	)
	ModifierRE = rex.New(modifierRE).MustCompile()
	// match emote name
	EmoteRE = rex.Common.Class(
		rex.Chars.Alphanumeric(),
		rex.Chars.Runes("-_():"),
	).Repeat().Between(2, 99)
	FloatRE = rex.Common.Raw(`[0-9]+\.?[0-9]*`)
	TokenRE = rex.New(rex.Group.Composite(
		EmoteRE,
		FloatRE,
		modifierRE,
	)).MustCompile()
)

type token interface {
	isToken()
}

type seventvEmoteToken string

func (seventvEmoteToken) isToken() {}

type intermediateEmoteToken string

func (intermediateEmoteToken) isToken() {}

type modifierToken string

func (modifierToken) isToken() {}

type floatToken float64

func (floatToken) isToken() {}

func hash(request string) string {
	h := sha1.New()
	h.Write([]byte(request))
	res := hex.EncodeToString(h.Sum(nil))
	log.Println(res, "IS", request)
	return res
}

type stack struct {
	elems []token
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

func (s *stack) push(elem token) {
	if s.err != nil {
		return
	}

	s.elems = append(s.elems, elem)
}

func (s *stack) popNum() float64 {
	if s.err != nil {
		return 0
	}

	if len(s.elems) == 0 {
		s.err = errors.New("tried to pop number from empty stack")
		return 0
	}

	switch res := s.pop().(type) {
	case floatToken:
		return float64(res)
	default:
		xerr.AppendInto(&s.err, xerr.New(
			xerr.WithMessage("tried to pop num"),
			xerr.WithField("got", res),
		))
		return 0
	}
}

func (s *stack) pop() token {
	if s.err != nil {
		return nil
	}

	if len(s.elems) == 0 {
		s.err = errors.New("tried to pop from empty stack")
		return nil
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

func emote(userID string, emoteToken token) (*webp.Animation, error) {
	var emotesRepo repository.EmotesRepository
	var seventvRepo seventv.Repository

	var emoteID string
	switch tok := emoteToken.(type) {
	case intermediateEmoteToken:
		emoteID = string(tok)
	case seventvEmoteToken:
		var err error
		emoteID, err = seventvRepo.GetEmoteID(userID, string(tok))
		if err != nil {
			return nil, err
		}

		if !emotesRepo.IsCached(string(tok)) {
			data, err := seventvRepo.Download7tvEmote(emoteID, emoteID)
			if err != nil {
				return nil, err
			}

			if err := emotesRepo.Save(data, emoteID); err != nil {
				return nil, err
			}
		}
	default:
		return nil, xerr.New(
			xerr.WithMessage("unexpected emote token type"),
			xerr.WithField("token", tok),
		)
	}

	return emotesRepo.LoadObject(emoteID)
}

func linearTokenHandler(
	userID string,
	stack *stack,
	modifierName modifierToken,
	construct func(float64, *webp.Animation) modifiers.Modifier,
) error {
	var emotesRepo repository.EmotesRepository

	// TODO: check stack errors after all pops
	arg := stack.pop()
	coeff := stack.popNum()

	defer bench(fmt.Sprintf("%s %s", modifierName, arg))()

	newEmote := hash(fmt.Sprintf("%s%s%f", arg, modifierName, coeff))
	if !emotesRepo.IsCached(newEmote) {
		img, err := emote(userID, arg)
		if err != nil {
			return err
		}

		m := construct(coeff, img)

		enc, err := m.Modify()
		if err != nil {
			return err
		}

		if err := emotesRepo.SaveObject(enc, newEmote); err != nil {
			return err
		}
	}

	stack.push(intermediateEmoteToken(newEmote))

	return nil
}

func unaryTokenHandler(
	userID string,
	stack *stack,
	modifierName modifierToken,
	construct func(*webp.Animation) modifiers.Modifier,
) error {
	var emotesRepo repository.EmotesRepository

	popped := stack.pop()
	var stackItem string
	switch arg := popped.(type) {
	case seventvEmoteToken:
		stackItem = string(arg)
	case intermediateEmoteToken:
		stackItem = string(arg)
	default:
		return xerr.New(
			xerr.WithMessage("expected emote on stack"),
			xerr.WithField("got", arg),
		)
	}

	defer bench(fmt.Sprintf("%s %s", modifierName, stackItem))()

	newEmote := hash(stackItem + string(modifierName))
	if !emotesRepo.IsCached(newEmote) {
		img, err := emote(userID, popped)
		if err != nil {
			return err
		}

		m := construct(img)

		enc, err := m.Modify()
		if err != nil {
			return xerr.New(
				xerr.WithErr(err),
				xerr.WithField("modifier", modifierName),
			)
		}

		if err := emotesRepo.SaveObject(enc, newEmote); err != nil {
			return err
		}
	}

	stack.push(intermediateEmoteToken(newEmote))

	return nil
}

func binaryTokenHandler(
	userID string,
	stack *stack,
	modifierName modifierToken,
	construct func(a, b *webp.Animation) modifiers.Modifier,
) error {
	var emotesRepo repository.EmotesRepository

	// TODO: assert stack size
	second := stack.pop()
	first := stack.pop()

	defer bench(fmt.Sprintf("%s %s %s", modifierName, first, second))()

	firstImg, err := emote(userID, first)
	if err != nil {
		return err
	}

	secondImg, err := emote(userID, second)
	if err != nil {
		return err
	}

	newEmote := hash(fmt.Sprintf("%s,%s%s", first, second, modifierName))
	if !emotesRepo.IsCached(newEmote) {
		m := construct(firstImg, secondImg)

		enc, err := m.Modify()
		if err != nil {
			return err
		}

		// TODO: do not re-evaluate if already exists (cache)
		if err := emotesRepo.SaveObject(enc, newEmote); err != nil {
			return err
		}
	}

	stack.push(intermediateEmoteToken(newEmote))

	return nil
}

func ParseTokens(tokens []string) []token {
	res := make([]token, len(tokens))
	for i, token := range tokens {
		if ModifierRE.MatchString(token) {
			res[i] = modifierToken(token)
		} else if f, err := strconv.ParseFloat(token, 64); err == nil {
			res[i] = floatToken(f)
		} else {
			res[i] = seventvEmoteToken(token)
		}
	}
	return res
}

func ProcessQuery(userID string, tokens []token) (intermediateEmoteToken, error) {
	stack := newStack()
	// TODO: assert all characters are used in tokenizing
	for _, rawToken := range tokens {
		switch token := rawToken.(type) {
		case seventvEmoteToken, intermediateEmoteToken:
			stack.push(token)
		case modifierToken:
			switch token {
			case ",":
			case ">revx":
				if err := unaryTokenHandler(
					userID,
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
					userID,
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
					userID,
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
					userID,
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
					userID,
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
					userID,
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
					userID,
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
					userID,
					&stack,
					token,
					func(coeff float64, x *webp.Animation) modifiers.Modifier {
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
					userID,
					&stack,
					token,
					func(coeff float64, x *webp.Animation) modifiers.Modifier {
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
					userID,
					&stack,
					token,
					func(coeff float64, x *webp.Animation) modifiers.Modifier {
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
					userID,
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
					userID,
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
					userID,
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
					userID,
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
					userID,
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
					userID,
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
					userID,
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
					userID,
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
					userID,
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
					userID,
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
				return "", xerr.New(
					xerr.WithMessage("unknown modifier"),
					xerr.WithField("token", token),
				)
			}
		default:
			return "", xerr.New(
				xerr.WithMessage("unknown token type"),
				xerr.WithField("token", token),
			)
		}
	}

	if err := stack.Err(); err != nil {
		return "", err
	}

	if stack.len() != 1 {
		return "", fmt.Errorf("stack has more than single item (or none): %v", stack)
	}

	top := stack.pop()
	res, ok := top.(intermediateEmoteToken)
	if !ok {
		return "", xerr.New(
			xerr.WithMessage("result is not intermediate token"),
			xerr.WithField("result", top),
		)
	}

	return res, nil
}
