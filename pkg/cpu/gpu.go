package cpu

import (
	"sort"

	"github.com/tbtommyb/goboy/pkg/constants"
	c "github.com/tbtommyb/goboy/pkg/constants"
)

const SpriteDataStartAddress = 0x8000

type GPU struct {
	cpu                *CPU
	display            DisplayInterface
	cyclesCounter      uint
	vBlankCounter      uint
	oams               []*oamEntry
	bgPixelVisibility  [constants.ScreenWidth]pixelVisibility
	interruptTriggered bool
	vram               [0x2000]byte
	sram               [0x100]byte
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
		cpu:  cpu,
		vram: [0x2000]byte{},
		sram: [0x100]byte{},
	}
	gpu.setStatusMode(SearchingOAMMode)
	return gpu
}

func (gpu *GPU) update() {
	if gpu.cpu.stop {
		return
	}
	if !gpu.getControl().isDisplayEnabled() {
		return
	}

	currentLine := gpu.cpu.ReadIO(c.LYAddress)
	currentMode := gpu.getStatus().mode()

	gpu.cyclesCounter++

	switch gpu.cyclesCounter {
	case 4:
		if currentMode != VBlankMode {
			gpu.setStatusMode(SearchingOAMMode)
			gpu.handleInterrupts()
		}
	case 80:
		if currentMode == SearchingOAMMode {
			gpu.parseOAMForScanline(currentLine)
			gpu.setStatusMode(TransferringMode)
		}
	case 252:
		if currentMode == TransferringMode {
			gpu.setStatusMode(HBlankMode)
			gpu.renderScanline(currentLine)
			gpu.handleInterrupts()
		}
	case CyclesPerScanline:
		newScanline := gpu.incrementScanline()
		gpu.cyclesCounter = 0

		if newScanline == VBlankStartScanline && currentMode != VBlankMode {
			gpu.setStatusMode(VBlankMode)
			gpu.requestInterrupt(VBlank)
			gpu.handleInterrupts()
		}
	}

	if gpu.getStatus().mode() == VBlankMode {
		gpu.vBlankCounter++
		if gpu.vBlankCounter == CyclesPerScanline*ScanlinesPerVBlank {
			gpu.setStatusMode(HBlankMode)
			gpu.resetScanline()
			gpu.cyclesCounter = 0
			gpu.vBlankCounter = 0
		}
		gpu.handleInterrupts()
	}

	if gpu.cpu.ReadIO(c.LYAddress) == gpu.cpu.ReadIO(c.LYCAddress) {
		gpu.setMatchFlag()
	} else {
		gpu.resetMatchFlag()
	}
}

func (gpu *GPU) handleInterrupts() {
	currentMode := gpu.getStatus().mode()

	if gpu.isModeInterruptSet(currentMode) || (gpu.isStatusSet(MatchFlag) && gpu.isStatusSet(MatchInterrupt)) {
		if !gpu.interruptTriggered {
			gpu.interruptTriggered = true
			gpu.cpu.requestInterrupt(LCDCStatus)
			return
		}
	}
	gpu.interruptTriggered = false
}

func (gpu *GPU) renderScanline(scanline byte) {
	control := gpu.getControl()
	for i := 0; i < constants.ScreenWidth; i++ {
		gpu.bgPixelVisibility[i] = invisible
	}

	if control.isBGEnabled() {
		gpu.renderBackground(scanline)
	}

	if control.isWindowEnabled() && scanline >= gpu.cpu.ReadIO(c.WindowYAddress) {
		gpu.renderWindow(scanline)
	}

	if control.isSpriteEnabled() {
		gpu.renderSprites(gpu.oams, scanline)
	}
}

func (gpu *GPU) resetScanline() {
	gpu.cpu.WriteIO(c.LYAddress, 0)
}

func (gpu *GPU) incrementScanline() byte {
	currentScanline := gpu.cpu.ReadIO(c.LYAddress)
	currentScanline++
	gpu.cpu.WriteIO(c.LYAddress, currentScanline)
	return currentScanline
}

func (gpu *GPU) requestInterrupt(interrupt Interrupt) {
	gpu.cpu.requestInterrupt(interrupt)
}

func (gpu *GPU) renderBackground(scanline byte) {
	scrollX := gpu.cpu.ReadIO(c.ScrollXAddress)
	scrollY := gpu.cpu.ReadIO(c.ScrollYAddress)

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
	winY := scanline - gpu.cpu.ReadIO(c.WindowYAddress)
	winStartX := int(gpu.cpu.ReadIO(c.WindowXAddress)) - int(ScrollXOffset)

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
	vramAddress := tileAddress - 0x8000
	return tileNum(gpu.vram[vramAddress])
}

func (gpu *GPU) fetchCharCodeBytes(baseAddress, tileOffset uint16) (byte, byte) {
	charCodeAddress := baseAddress + (uint16(tileOffset) << 1)
	vramAddress := charCodeAddress - 0x8000
	low := gpu.vram[vramAddress]
	high := gpu.vram[vramAddress+1]
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
	vramAddress := spriteAddress - 0x8000
	low := gpu.vram[vramAddress+(charCode<<1)]
	high := gpu.vram[vramAddress+(charCode<<1)+1]
	return low, high
}

func getColourCodeFrom(xPos, low, high byte) colourCode {
	bitOffset := xPos & CharCodeMask
	bitL := (low >> (7 - bitOffset)) & 0x1
	bitH := (high >> (7 - bitOffset)) & 0x1
	return colourCode((bitH << 1) | bitL)
}

func (gpu *GPU) applyBGPalette(colour colourCode) RGB {
	paletteRegister := gpu.cpu.ReadIO(c.BGPAddress)
	return applyPalette(selectColourCode(paletteRegister, colour))
}

func (gpu *GPU) applySpritePalette(colour colourCode, e *oamEntry) RGB {
	paletteRegister := gpu.cpu.ReadIO(c.OBP0Address)
	if e.useOBP1() {
		paletteRegister = gpu.cpu.ReadIO(c.OBP1Address)
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
	bytes := gpu.sram[location : location+ObjDataSize]
	return int16(bytes[0]) - SpriteYOffset, int16(bytes[1]) - SpriteXOffset, tileNum(bytes[2]), flags(bytes[3])
}

type sortableOAM []*oamEntry

func (s sortableOAM) Less(i, j int) bool { return s[i].x < s[j].x }
func (s sortableOAM) Len() int           { return len(s) }
func (s sortableOAM) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (gpu *GPU) writeOAM(addr uint16, val byte) {
	currentMode := gpu.getStatus().mode()
	if !(currentMode == SearchingOAMMode || currentMode == TransferringMode) {
		gpu.sram[addr-0xFE00] = val
	}
}

func (gpu *GPU) readOAM(addr uint16) byte {
	currentMode := gpu.getStatus().mode()
	if !(currentMode == SearchingOAMMode || currentMode == TransferringMode) {
		return gpu.sram[addr]
	}
	return 0xff
}

func (gpu *GPU) writeVRAM(addr uint16, val byte) {
	currentMode := gpu.getStatus().mode()
	if currentMode != TransferringMode {
		gpu.vram[addr-0x8000] = val
	}
}

func (gpu *GPU) readVRAM(addr uint16) byte {
	currentMode := gpu.getStatus().mode()
	if currentMode != TransferringMode {
		return gpu.vram[addr-0x8000]
	}
	return 0xff
}
