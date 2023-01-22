# twitch emotes modifier

Modify and combine twitch emotes to create new ones.

## How to use

This project defines stack language used like this:
```
62c5c34724fb1819d9f08b4d>dup>stacky,63053d5701b7faa7ef5f2a78>iscaley>iscalex>stackx
```
This program does following:
1. pushes [62c5c34724fb1819d9f08b4d](https://7tv.app/emotes/62c5c34724fb1819d9f08b4d) emote on stack
1. `dup`licates emote on stack
1. `stack`s two emotes on stack by `y` axis, then pushes result back to stack
1. pushes [63053d5701b7faa7ef5f2a78](https://7tv.app/emotes/63053d5701b7faa7ef5f2a78) emote on stack
1. `i`ncreases `scale` of top stack emote by `y` axis, then pushes result back to stack
1. `i`ncreases `scale` of top stack emote by `x` axis, then pushes result back to stack
1. `stack`s two emotes on stack by `x` axis, then pushes result back to stack

(someday i will make more graphicals explanations)

### Examples
Some examples to begin with
```
603caea243b9e100141caf4f,6129ea8afd97806f9d734a76>over
60b0c36388e8246a4b120d7e>dup>revx>stackx>dup>revy>stacky>revx
611523959bf574f1fded6d72>dscalex
611523959bf574f1fded6d72>dscaley
611523959bf574f1fded6d72>iscalex>iscalex>iscalex>dscalet>dscalet
62c5c34724fb1819d9f08b4d>dup>stacky,63053d5701b7faa7ef5f2a78>iscaley>iscalex>stackx
62c5c34724fb1819d9f08b4d,63053d5701b7faa7ef5f2a78,62c5c34724fb1819d9f08b4d>stackx>stackx
63053d5701b7faa7ef5f2a78,62c5c34724fb1819d9f08b4d>stackx
63053d5701b7faa7ef5f2a78>iscaley>iscalex
```

## Available modifiers

### Meta modifiers
Meta modifiers just manipulates stack.
- `>dup` - `dup`licates last item on stack
- `>swap` - `swap`s two last items on stack

### Unary modifiers
Unary modifiers take emote from stack, modifies it, then pushes result back to stack.
- `>revx` - `rev`erses emote by `x` axis
- `>revy` - `rev`erses emote by `y` axis
- `>revt` - `rev`erses emote by `t`ime axis
- `>iscalex` - `i`ncreases `scale` of emote by `x` axis
- `>iscaley` - `i`ncreases `scale` of emote by `y` axis
- `>iscalet` - `i`ncreases `scale` of emote by `t`ime axis, slowing it down
- `>dscalex` - `d`ecreases `scale` of emote by `x` axis
- `>dscaley` - `d`ecreases `scale` of emote by `y` axis
- `>dscalet` - `d`ecreases `scale` of emote by `t`ime axis, speeding it up

### Binary modifiers
Binary modifiers take two emotes from stack, modifies them into single emote, then pushes result back to stack.
- `>over` - puts one emote `over` another
- `>stackx` - `stack`s emotes by `x` axis
- `>stacky` - `stack`s emotes by `y` axis
- `>stackt` - `stack`s emotes by `t`ime axis, so one emote is showed first, then other

## Known limitations
- images must match exactly by size, e.g. when `over` - they must be exactly of same size, when `stackx` - must be same height, etc.
- merged animated emotes may produce somehow laggy animations with frames longer than they must be. Not sure why, because of go bindings for `libwebp`, because of `libwebp`, because of format limitations or what
- very early version, e.g. trying to stack when there are no items to stack will lead to cryptic message, so use carefully
- not optimised in any way, larger emotes or emotes with high number of frames takes proportionally longer time to process

