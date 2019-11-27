# Goboy

A Gameboy emulator written in Go. Follow progress [here](https://tmjohnson.co.uk/tags/goboy/).

## Running

```go
go get github.com/hajimehoshi/ebiten
go run cmd/goboy/main.go YOUR_ROM_HERE
```

If you have a BIOS/bootloader ROM you can get the scrolling Nintendo logo by specifying the ROM e.g.:

```go
go run cmd/goboy/main.go -bios bios.gb mario.gb
```

Builds coming soon.

I have tested with Tetris, Zelda, Kirby and Super Mario World. All work so far. Audio needs implemented.  There is some flickering I haven't had time to investigate yet.

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
