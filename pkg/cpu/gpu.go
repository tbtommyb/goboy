package cpu

import (
	"sort"

	"github.com/tbtommyb/goboy/pkg/constants"
)

type GPU struct {
	cpu               *CPU
	display           DisplayInterface
	cyclesCounter     uint
	oams              []*oamEntry
	bgPixelVisibility [constants.ScreenWidth]pixelVisibility
}

type DisplayInterface interface {
	WritePixel(x, y, r, g, b byte)
}

type RGB struct {
	r, g, b byte
}

type tileNum byte
type colourCode byte
type flags byte

type Mode byte

const (
	HBlankMode       Mode = 0
	VBlankMode            = 1
	SearchingOAMMode      = 2
	TransferringMode      = 3
)

type addressingMode bool

const (
	signedAddressing   = true
	unsignedAddressing = false
)

type pixelVisibility bool

const (
	visible   = true
	invisible = false
)

// Cycle counts taken from Pandocs
const (
	CyclesPerScanline         uint = 456
	CyclesPerSearchingOAMMode      = 80
	CyclesPerTransferringMode      = 172
	CyclesPerHBlankMode            = 204
)

const (
	MaxScanline                byte = 153
	VBlankStartScanline             = 144
	ScanlinesPerVBlank              = 10
	SearchingOAMModeCycleBound      = 376
	TransferringModeCycleBound      = 302
	StatusModeResetMask             = 0xFC
	ModeMask                        = 3
)

const (
	ScrollXOffset         byte = 7
	CharCodeMask               = 7
	CharCodeShift              = 1
	CharCodeSize               = 16
	TilePixelSize              = 8
	TileRowSize                = 32
	BitsPerColour              = 2
	ColourMask                 = 3
	SpriteDataSize             = 16
	SpritePixelSize            = 8
	DoubleSpriteHeight         = 16
	ObjCount                   = 128
	ObjDataSize                = 4
	MaxSpritesPerScanline      = 10
	MaxSpritesPerScreen        = 40
	SpriteByteSize             = 4
	SpriteYOffset              = 16
	SpriteXOffset              = 8
)

var standardPalette = []byte{0xff, 0xaa, 0x55, 0x00}

func InitGPU(cpu *CPU) *GPU {
	gpu := &GPU{
		cpu: cpu,
	}
	gpu.setMode(SearchingOAMMode)
	return gpu
}

func (gpu *GPU) setMode(mode Mode) {
	gpu.setStatus(gpu.getStatus().setMode(mode))
}

func (gpu *GPU) update(cycles uint) {
	control := gpu.getControl()
	if !control.isDisplayEnabled() {
		gpu.resetScanline()
		gpu.setMode(VBlankMode)
		return
	}

	currentLine := gpu.cpu.getLY()
	status := gpu.getStatus()
	currentMode := status.mode()
	newMode := currentMode

	gpu.cyclesCounter += cycles

	switch currentMode {
	case SearchingOAMMode:
		if gpu.cyclesCounter >= CyclesPerSearchingOAMMode {
			gpu.cyclesCounter = 0
			newMode = TransferringMode
			gpu.parseOAMForScanline(currentLine)
		}
	case TransferringMode:
		if gpu.cyclesCounter >= CyclesPerTransferringMode {
			gpu.cyclesCounter = 0
			newMode = HBlankMode
			gpu.renderScanline(currentLine)
		}
	case HBlankMode:
		if gpu.cyclesCounter >= CyclesPerHBlankMode {
			gpu.cyclesCounter = 0
			newScanline := gpu.incrementScanline()

			if newScanline == VBlankStartScanline {
				newMode = VBlankMode
				gpu.requestInterrupt(VBlank)
			} else {
				newMode = SearchingOAMMode
			}
		}
	case VBlankMode:
		if gpu.cyclesCounter >= CyclesPerScanline {
			gpu.cyclesCounter = 0
			newScanline := gpu.incrementScanline()
			// TODO: refactor to put wrapping logic in incrementScanline
			if newScanline > MaxScanline { // TODO -1 or not?
				newMode = SearchingOAMMode
				gpu.resetScanline()
			}
		}
	}

	status = status.setMode(newMode)
	if status.isModeInterruptSet() && (newMode != currentMode) {
		gpu.cpu.requestInterrupt(LCDCStatus)
	}
	if gpu.cpu.getLY() == gpu.cpu.getLYC() {
		status = status.setMatchFlag()
		if status.isMatchInterruptSet() {
			gpu.cpu.requestInterrupt(LCDCStatus)

		}
	} else {
		status = status.resetMatchFlag()
	}

	gpu.setStatus(status)

}

func (gpu *GPU) renderScanline(scanline byte) {
	control := gpu.getControl()
	for i := 0; i < constants.ScreenWidth; i++ {
		gpu.bgPixelVisibility[i] = invisible
	}

	gpu.renderBackground(scanline)

	if control.isWindowEnabled() && scanline >= gpu.cpu.getWindowY() {
		gpu.renderWindow(scanline)
	}

	if control.isSpriteEnabled() {
		gpu.renderSprites(gpu.oams, scanline)
	}
}

func (gpu *GPU) resetScanline() {
	gpu.cpu.setLY(0)
}

func (gpu *GPU) incrementScanline() byte {
	currentScanline := gpu.cpu.getLY()
	currentScanline++
	gpu.cpu.setLY(currentScanline)
	return currentScanline
}

func (gpu *GPU) requestInterrupt(interrupt Interrupt) {
	gpu.cpu.requestInterrupt(interrupt)
}

func (gpu *GPU) renderBackground(scanline byte) {
	scrollX := gpu.cpu.getScrollX()
	scrollY := gpu.cpu.getScrollY()

	startAddress := gpu.bgTileMapStartAddress()

	yPos := byte(scrollY + scanline)

	for x := byte(0); x < byte(constants.ScreenWidth); x++ {
		xPos := byte(scrollX + x)

		colour := gpu.fetchBackgroundColour(startAddress, xPos, yPos)
		if colour != 0 {
			gpu.bgPixelVisibility[x] = visible
		}

		rgb := gpu.applyBGPalette(colour)
		gpu.display.WritePixel(x, scanline, rgb.r, rgb.g, rgb.b)
	}
}

func (gpu *GPU) renderSprites(oams []*oamEntry, scanline byte) {
	for _, e := range oams {
		var startX byte = 0
		if e.x > 0 {
			startX = byte(e.x)
		}
		endX := byte(e.x + SpritePixelSize)

		for x := startX; x < endX && x < byte(constants.ScreenWidth); x++ {
			if e.behindBG() && gpu.bgPixelVisibility[x] == visible {
				continue
			}
			if rgb, isVisible := gpu.fetchSpritePixel(e, x, scanline); isVisible {
				gpu.display.WritePixel(x, scanline, rgb.r, rgb.r, rgb.b)
			}
		}
	}
}

func (gpu *GPU) renderWindow(scanline byte) {
	winY := scanline - gpu.cpu.getWindowY()
	winStartX := int(gpu.cpu.getWindowX()) - int(ScrollXOffset)

	for x := winStartX; x < constants.ScreenWidth; x++ {
		if x < 0 {
			continue
		}

		startAddress := gpu.windowTileMapStartAddress()
		colour := gpu.fetchBackgroundColour(startAddress, byte(x-winStartX), winY)
		if colour != 0 {
			gpu.bgPixelVisibility[x] = true
		}

		rgb := gpu.applyBGPalette(colour)
		gpu.display.WritePixel(byte(x), scanline, rgb.r, rgb.g, rgb.b)
	}
}

func (gpu *GPU) windowTileMapStartAddress() uint16 {
	if gpu.getControl().useHighWindowAddress() {
		return 0x9C00
	}
	return 0x9800
}

func (gpu *GPU) bgTileDataAddress() (uint16, addressingMode) {
	if gpu.getControl().useHighBGDataAddress() {
		return 0x8800, signedAddressing
	}
	return 0x8000, unsignedAddressing
}

func (gpu *GPU) bgTileMapStartAddress() uint16 {
	if gpu.getControl().useHighBGStartAddress() {
		return 0x9C00
	}
	return 0x9800
}

func getTileLocation(addressMode addressingMode, baseAddress uint16, tile tileNum) uint16 {
	if addressMode == signedAddressing {
		tile = tileNum(int(tile) + ObjCount)
	}
	return baseAddress + (uint16(tile) * CharCodeSize)
}

func (gpu *GPU) fetchBackgroundColour(startAddress uint16, x, y byte) colourCode {
	dataAddress, addressMode := gpu.bgTileDataAddress()
	tileNum := gpu.fetchTileNum(startAddress, x, y)
	tileLocation := getTileLocation(addressMode, dataAddress, tileNum)
	charCode := y & CharCodeMask
	low, high := gpu.fetchCharCodeBytes(tileLocation, uint16(charCode))
	return getColourCodeFrom(x, low, high)
}

func (gpu *GPU) fetchTileNum(startAddress uint16, xPos, yPos byte) tileNum {
	tileNumX, tileNumY := uint16(xPos/TilePixelSize), uint16(yPos/TilePixelSize)
	tileAddress := uint16(startAddress + tileNumY*TileRowSize + tileNumX)
	return tileNum(gpu.cpu.memory.get(tileAddress))
}

func (gpu *GPU) fetchCharCodeBytes(baseAddress, tileOffset uint16) (byte, byte) {
	charCodeAddress := baseAddress + (uint16(tileOffset) << 1)
	low := gpu.cpu.memory.get(charCodeAddress)
	high := gpu.cpu.memory.get(charCodeAddress + 1)
	return low, high
}

func (gpu *GPU) fetchSpritePixel(e *oamEntry, x, y byte) (RGB, pixelVisibility) {
	tileX := byte(int16(x) - e.x)
	tileY := byte(int16(y) - e.y)

	if e.xFlip() {
		tileX = 7 - tileX
	}
	if e.yFlip() {
		tileY = e.height - 1 - tileY
	}
	tile := e.tileNum
	if e.height == DoubleSpriteHeight {
		tile = tileNum(ignoreLowerBit(byte(tile)))
		if tileY >= SpritePixelSize {
			tile++
		}
	}

	charCode := uint16(tileY & CharCodeMask)
	spriteAddress := getSpriteAddress(tile)
	low, high := gpu.fetchSpriteData(spriteAddress, charCode)
	colour := getColourCodeFrom(tileX, low, high)
	if colour == 0 {
		return RGB{0, 0, 0}, invisible
	}

	rgb := gpu.applySpritePalette(colour, e)
	return rgb, visible
}

func ignoreLowerBit(val byte) byte {
	return val &^ 0x1
}

func getSpriteAddress(tileNum tileNum) uint16 {
	return SpriteDataStartAddress + (uint16(tileNum) * SpriteDataSize)
}

func (gpu *GPU) fetchSpriteData(spriteAddress, charCode uint16) (byte, byte) {
	low := gpu.cpu.memory.get(spriteAddress + (charCode << 1))
	high := gpu.cpu.memory.get(spriteAddress + (charCode << 1) + 1)
	return low, high
}

func getColourCodeFrom(xPos, low, high byte) colourCode {
	bitOffset := xPos & CharCodeMask
	bitL := (low >> (7 - bitOffset)) & 0x1
	bitH := (high >> (7 - bitOffset)) & 0x1
	return colourCode((bitH << 1) | bitL)
}

func (gpu *GPU) applyBGPalette(colour colourCode) RGB {
	paletteRegister := gpu.cpu.getBGP()
	return applyPalette(selectColourCode(paletteRegister, colour))
}

func (gpu *GPU) applySpritePalette(colour colourCode, e *oamEntry) RGB {
	paletteRegister := gpu.cpu.getOBP0()
	if e.useOBP1() {
		paletteRegister = gpu.cpu.getOBP1()
	}
	return applyPalette(selectColourCode(paletteRegister, colour))
}

func selectColourCode(register byte, colour colourCode) colourCode {
	return colourCode((register >> (colour * BitsPerColour)) & ColourMask)
}

func applyPalette(code colourCode) RGB {
	colour := standardPalette[code]
	return RGB{r: colour, g: colour, b: colour}
}

type oamEntry struct {
	x       int16
	y       int16
	height  byte
	tileNum tileNum
	flags   flags
}

func (e *oamEntry) behindBG() bool { return e.flags&0x80 != 0 }
func (e *oamEntry) yFlip() bool    { return e.flags&0x40 != 0 }
func (e *oamEntry) xFlip() bool    { return e.flags&0x20 != 0 }
func (e *oamEntry) useOBP1() bool  { return e.flags&0x10 != 0 }

func yInSprite(scanline byte, y int16, height byte) bool {
	return int16(scanline) >= y && int16(scanline) < y+int16(height)
}

func (gpu *GPU) parseOAMForScanline(scanline byte) {
	// This method accesses memory directly to avoid mode restrictions
	height := byte(SpritePixelSize)
	if gpu.getControl().useBigSprites() {
		height = 16
	}
	gpu.oams = gpu.oams[:0]
	for i := 0; len(gpu.oams) < MaxSpritesPerScanline && i < MaxSpritesPerScreen; i++ {
		offset := uint16(i * SpriteByteSize)
		y, x, num, flags := gpu.fetchOAMData(offset)
		if !yInSprite(scanline, y, height) {
			continue
		}
		gpu.oams = append(gpu.oams, &oamEntry{
			x:       x,
			y:       y,
			height:  height,
			tileNum: num,
			flags:   flags,
		})
	}

	sort.Stable(sortableOAM(gpu.oams))
}

func (gpu *GPU) fetchOAMData(location uint16) (int16, int16, tileNum, flags) {
	bytes := gpu.cpu.memory.sram[location : location+ObjDataSize]
	return int16(bytes[0]) - SpriteYOffset, int16(bytes[1]) - SpriteXOffset, tileNum(bytes[2]), flags(bytes[3])
}

type sortableOAM []*oamEntry

func (s sortableOAM) Less(i, j int) bool { return s[i].x < s[j].x }
func (s sortableOAM) Len() int           { return len(s) }
func (s sortableOAM) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (gpu *GPU) writeOAM(addr uint16, val byte) {
	currentMode := gpu.getStatus().mode()
	if !(currentMode == SearchingOAMMode || currentMode == TransferringMode) {
		gpu.cpu.memory.sram[addr-0xFE00] = val
	}
}

func (gpu *GPU) readOAM(addr uint16) byte {
	currentMode := gpu.getStatus().mode()
	if !(currentMode == SearchingOAMMode || currentMode == TransferringMode) {
		return gpu.cpu.memory.sram[addr-0xFE00]
	}
	return 0xff
}
