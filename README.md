# Goboy

A Gameboy emulator written in Go. Follow progress [here](https://tmjohnson.co.uk/tags/goboy/).

## Running

```go
go get github.com/hajimehoshi/ebiten
go run cmd/goboy/main.go -rom YOUR_ROM_HERE
```

I have tested with Tetris and Super Mario World. Both work so far though with lots of bugs e.g:
- flickering / corrupted sprites
- you need to hit START (enter) for the game to load
- you seem to need to hit A/B a few times before START works on the title screen.

## Buttons

|Gameboy|Keyboard|
|---|---|
|start|enter|
|select|backspace|
|up|up arrow|
|down|down arrow|
|left|left arrow|
|right|right arrow|
|A|Z|
|B|X|
