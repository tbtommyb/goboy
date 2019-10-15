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
	backgroundVisible [constants.ScreenWidth]bool
}

type DisplayInterface interface {
	WritePixel(x, y, r, g, b byte)
}

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

type visibility bool

const (
	invisible = false
	visible   = true
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
	MaxVisibleScanline              = 144
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
	ColourSize                 = 2
	ColourMask                 = 3
	SpriteDataSize             = 16
	SpritePixelSize            = 8
	DoubleSpriteHeight         = 16
	ObjCount                   = 128
	MaxSpritesPerScanline      = 10
	MaxSpritesPerScreen        = 40
	SpriteByteSize             = 4
	SpriteYOffset              = 16
	SpriteXOffset              = 8
)

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

func (gpu *GPU) update() {
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

	gpu.cyclesCounter++

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
			gpu.incrementScanline()

			if currentLine == MaxVisibleScanline-1 {
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
			if newScanline > MaxScanline-1 {
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
		gpu.backgroundVisible[i] = false
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

func (gpu *GPU) renderBackground(scanline byte) {
	scrollX := gpu.cpu.getScrollX()
	scrollY := gpu.cpu.getScrollY()

	startAddress := gpu.bgTileMapStartAddress()

	yPos := byte(scrollY + scanline)

	for x := byte(0); x < byte(constants.ScreenWidth); x++ {
		xPos := byte(scrollX + x)

		pixel := gpu.getBackgroundPixel(startAddress, xPos, yPos)
		if pixel != 0 {
			gpu.backgroundVisible[x] = true
		}

		r, g, b := gpu.applyBGPalette(pixel)
		gpu.display.WritePixel(x, scanline, r, g, b)
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
			if e.behindBG() && gpu.backgroundVisible[x] {
				continue
			}
			if r, g, b, isVisible := gpu.getSpritePixel(e, x, scanline); isVisible {
				gpu.display.WritePixel(x, scanline, r, g, b)
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
		pixel := gpu.getBackgroundPixel(startAddress, byte(x-winStartX), winY)
		if pixel != 0 {
			gpu.backgroundVisible[x] = true
		}

		r, g, b := gpu.applyBGPalette(pixel)
		gpu.display.WritePixel(byte(x), scanline, r, g, b)
	}
}

func (gpu *GPU) requestInterrupt(interrupt Interrupt) {
	gpu.cpu.requestInterrupt(interrupt)
}

func (gpu *GPU) windowTileMapStartAddress() uint16 {
	if gpu.getControl().isHighWindowAddress() {
		return 0x9C00
	}
	return 0x9800
}

func (gpu *GPU) bgTileDataAddress() (uint16, addressingMode) {
	if gpu.getControl().isHighBGDataAddress() {
		return 0x8800, signedAddressing
	}
	return 0x8000, unsignedAddressing
}

func (gpu *GPU) bgTileMapStartAddress() uint16 {
	if gpu.getControl().isHighBGStartAddress() {
		return 0x9C00
	}
	return 0x9800
}

func getTileLocation(addressMode addressingMode, baseAddress, tileNum uint16) uint16 {
	if addressMode == signedAddressing {
		tileNum = uint16(int(int8(tileNum)) + ObjCount)
	}
	return baseAddress + (tileNum * CharCodeSize)
}

func (gpu *GPU) getBackgroundPixel(startAddress uint16, x, y byte) byte {
	dataAddress, addressMode := gpu.bgTileDataAddress()
	tileNum := gpu.getTileNum(startAddress, x, y)
	tileLocation := getTileLocation(addressMode, dataAddress, tileNum)
	charCode := y & CharCodeMask
	low, high := gpu.fetchCharCodeBytes(tileLocation, uint16(charCode))
	return fetchBitPair(x, low, high)
}

func (gpu *GPU) getTileNum(startAddress uint16, xPos, yPos byte) uint16 {
	tileNumX, tileNumY := uint16(xPos/TilePixelSize), uint16(yPos/TilePixelSize)
	tileAddress := uint16(startAddress + tileNumY*TileRowSize + tileNumX)
	return uint16(gpu.cpu.memory.get(tileAddress))
}

func (gpu *GPU) fetchCharCodeBytes(baseAddress, tileOffset uint16) (byte, byte) {
	charCodeAddress := baseAddress + (uint16(tileOffset) << 1)
	low := gpu.cpu.memory.get(charCodeAddress)
	high := gpu.cpu.memory.get(charCodeAddress + 1)
	return low, high
}

func (gpu *GPU) getSpritePixel(e *oamEntry, x, y byte) (byte, byte, byte, visibility) {
	tileX := byte(int16(x) - e.x)
	tileY := byte(int16(y) - e.y)

	if e.xFlip() {
		tileX = 7 - tileX
	}
	if e.yFlip() {
		tileY = e.height - 1 - tileY
	}
	tileNum := e.tileNum
	if e.height == DoubleSpriteHeight {
		tileNum = ignoreLowerBit(tileNum)
		if tileY >= SpritePixelSize {
			tileNum++
		}
	}

	charCode := uint16(tileY & CharCodeMask)
	spriteAddress := getSpriteAddress(tileNum)
	low, high := gpu.fetchSpriteData(spriteAddress, charCode)
	colour := fetchBitPair(tileX, low, high)
	if colour == 0 {
		return 0, 0, 0, invisible
	}

	palettedPixel := gpu.applySpritePalette(e, colour)
	r, g, b := applyCustomPalette(palettedPixel)
	return r, g, b, visible
}

func (gpu *GPU) fetchSpriteData(spriteAddress, charCode uint16) (byte, byte) {
	low := gpu.cpu.memory.get(spriteAddress + (charCode << 1))
	high := gpu.cpu.memory.get(spriteAddress + (charCode << 1) + 1)
	return low, high
}

func ignoreLowerBit(val byte) byte {
	return val &^ 0x1
}

func fetchBitPair(xPos, low, high byte) byte {
	bitOffset := xPos & CharCodeMask
	bitL := (low >> (7 - bitOffset)) & 0x1
	bitH := (high >> (7 - bitOffset)) & 0x1
	return (bitH << 1) | bitL
}

func getSpriteAddress(tileNum byte) uint16 {
	return CharacterStart + (uint16(tileNum) * SpriteDataSize)
}

var standardPalette = [][]byte{
	{0xff, 0xff, 0xff},
	{0xaa, 0xaa, 0xaa},
	{0x55, 0x55, 0x55},
	{0x00, 0x00, 0x00},
}

func applyCustomPalette(val byte) (byte, byte, byte) {
	outVal := standardPalette[val]
	return outVal[0], outVal[1], outVal[2]
}

func (gpu *GPU) applyBGPalette(colour byte) (byte, byte, byte) {
	customColour := (gpu.cpu.getBGP() >> (colour * ColourSize)) & ColourMask
	return applyCustomPalette(customColour)
}

func (gpu *GPU) applySpritePalette(e *oamEntry, colour byte) byte {
	palReg := gpu.cpu.getOBP0()
	if e.palSelector() {
		palReg = gpu.cpu.getOBP1()
	}

	return (palReg >> uint((colour * ColourSize))) & ColourMask
}

type oamEntry struct {
	x         int16
	y         int16
	height    byte
	tileNum   byte
	flagsByte byte
}

func (e *oamEntry) behindBG() bool    { return e.flagsByte&0x80 != 0 }
func (e *oamEntry) yFlip() bool       { return e.flagsByte&0x40 != 0 }
func (e *oamEntry) xFlip() bool       { return e.flagsByte&0x20 != 0 }
func (e *oamEntry) palSelector() bool { return e.flagsByte&0x10 != 0 }

func yInSprite(scanline byte, y int16, height int) bool {
	return int16(scanline) >= y && int16(scanline) < y+int16(height)
}

func (gpu *GPU) parseOAMForScanline(scanline byte) {
	// This method accesses memory directly to avoid mode restrictions
	gpu.oams = gpu.oams[:0]
	for i := 0; len(gpu.oams) < MaxSpritesPerScanline && i < MaxSpritesPerScreen; i++ {
		offset := uint16(i * SpriteByteSize)
		y, x, num, flags := gpu.fetchOAMData(offset)
		if !yInSprite(scanline, y, SpritePixelSize) {
			continue
		}
		gpu.oams = append(gpu.oams, &oamEntry{
			x:         x,
			y:         y,
			height:    SpritePixelSize,
			tileNum:   num,
			flagsByte: flags,
		})
	}

	sort.Stable(sortableOAM(gpu.oams))
}

func (gpu *GPU) fetchOAMData(location uint16) (int16, int16, byte, byte) {
	bytes := gpu.cpu.memory.sram[location : location+4]
	return int16(bytes[0]) - SpriteYOffset, int16(bytes[1]) - SpriteXOffset, bytes[2], bytes[3]
}

type sortableOAM []*oamEntry

func (s sortableOAM) Less(i, j int) bool { return s[i].x < s[j].x }
func (s sortableOAM) Len() int           { return len(s) }
func (s sortableOAM) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// -----------
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
