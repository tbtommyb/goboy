package cpu

import (
	"sort"

	"github.com/tbtommyb/goboy/pkg/constants"
	"github.com/tbtommyb/goboy/pkg/utils"
)

type Mode byte
type Status byte

type GPU struct {
	status        Status
	cpu           *CPU
	display       DisplayInterface
	cyclesCounter uint
	oams          []*oamEntry
	// for oam sprite priority
	BGMask     [constants.ScreenWidth]bool
	SpriteMask [constants.ScreenWidth]bool
}

type DisplayInterface interface {
	WritePixel(x, y, r, g, b, a byte)
}

const (
	MaxScanline                byte = 153
	MaxVisibleScanline              = 144
	ScanlinesPerVBlank              = 10
	SearchingOAMModeCycleBound      = 376
	TransferringModeCycleBound      = 302
	StatusModeResetMask             = 0xFC
	ModeMask                        = 3
	CyclesPerScanline          uint = 456
)

const (
	ScrollXOffset   byte = 7
	CharCodeMask         = 7
	CharCodeShift        = 1
	CharCodeSize         = 16
	TilePixelSize        = 8
	TileRowSize          = 32
	ColourSize           = 2
	ColourMask           = 3
	SpriteDataSize       = 16
	SpritePixelSize      = 8
	ObjCount             = 128
)

const (
	HBlankMode       Mode = 0
	VBlankMode            = 1
	SearchingOAMMode      = 2
	TransferringMode      = 3
)

var standardPalette = [][]byte{
	{0xff, 0xff, 0xff},
	{0xaa, 0xaa, 0xaa},
	{0x55, 0x55, 0x55},
	{0x00, 0x00, 0x00},
}

func InitGPU(cpu *CPU) *GPU {
	gpu := &GPU{
		cpu: cpu,
	}
	gpu.setMode(SearchingOAMMode)
	return gpu
}

func (gpu *GPU) update(_ uint) {
	if !gpu.cpu.isLCDCSet(LCDDisplayEnable) {
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
		if gpu.cyclesCounter >= 80 {
			gpu.cyclesCounter = 0
			newMode = TransferringMode
			gpu.parseOAMForScanline(currentLine)
		}
	case TransferringMode:
		if gpu.cyclesCounter >= 172 {
			gpu.cyclesCounter = 0
			newMode = HBlankMode
			gpu.renderScanline(currentLine)
		}
	case HBlankMode:
		if gpu.cyclesCounter >= 204 {
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
	for i := 0; i < constants.ScreenWidth; i++ {
		gpu.BGMask[i] = false
		gpu.SpriteMask[i] = false
	}
	gpu.renderBackground(scanline)
	if gpu.cpu.isLCDCSet(WindowDisplayEnable) && scanline >= gpu.cpu.getWindowY() {
		gpu.renderWindow(scanline)
	}

	if gpu.cpu.isLCDCSet(SpriteEnable) {
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

		pixel := gpu.getPixel(startAddress, xPos, yPos)
		if pixel != 0 {
			gpu.BGMask[x] = true
		}
		r, g, b := gpu.applyBGPalette(pixel)

		gpu.display.WritePixel(x, scanline, r, g, b, 0xff)
	}
}

func (gpu *GPU) renderSprites(oams []*oamEntry, scanline byte) {
	for _, e := range oams {
		startX := byte(0)
		if e.x > 0 {
			startX = byte(e.x)
		}
		endX := byte(e.x + SpritePixelSize)

		for x := startX; x < endX && x < byte(constants.ScreenWidth); x++ {
			// TODO: hide sprites
			hideSprite := e.behindBG() && gpu.BGMask[x]
			if !hideSprite && !gpu.SpriteMask[x] {
				if r, g, b, a := gpu.getSpritePixel(e, x, scanline); a {
					gpu.display.WritePixel(x, scanline, r, g, b, 0xff)
					gpu.SpriteMask[x] = true
				}
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
		pixel := gpu.getPixel(startAddress, byte(x-winStartX), winY)
		if pixel != 0 {
			gpu.BGMask[x] = true
		}
		r, g, b := gpu.applyBGPalette(pixel)
		gpu.display.WritePixel(byte(x), scanline, r, g, b, 0xff)
	}
}

func (gpu *GPU) requestInterrupt(interrupt Interrupt) {
	gpu.cpu.requestInterrupt(interrupt)
}

func (gpu *GPU) windowTileMapStartAddress() uint16 {
	if gpu.cpu.isLCDCSet(WindowTileMapDisplaySelect) {
		return 0x9C00
	}
	return 0x9800
}

func (gpu *GPU) bgTileDataAddress() (uint16, bool) {
	if gpu.cpu.isLCDCSet(DataSelect) {
		return 0x8000, true
	} else {
		return 0x8800, false
	}
}

func (gpu *GPU) bgTileMapStartAddress() uint16 {
	flag := LCDCFlag(BGTileMapDisplaySelect)
	if gpu.cpu.isLCDCSet(flag) {
		return 0x9C00
	}
	return 0x9800
}

func (gpu *GPU) applyCustomPalette(val byte) (byte, byte, byte) {
	outVal := standardPalette[val]
	return outVal[0], outVal[1], outVal[2]
}

func (gpu *GPU) getTileNum(startAddress uint16, xPos, yPos byte) uint16 {
	tileNumX, tileNumY := uint16(xPos/TilePixelSize), uint16(yPos/TilePixelSize)
	tileAddress := uint16(startAddress + tileNumY*TileRowSize + tileNumX)
	return uint16(gpu.cpu.memory.get(tileAddress))
}

func (gpu *GPU) getPixel(startAddress uint16, x, y byte) byte {
	dataAddress, unsig := gpu.bgTileDataAddress()
	tileNum := gpu.getTileNum(startAddress, x, y)
	tileLocation := getTileLocation(unsig, dataAddress, tileNum)
	charCode := y & CharCodeMask
	low, high := gpu.fetchCharCodeBytes(tileLocation, uint16(charCode))
	return gpu.fetchBitPair(x, low, high)
}

func getTileLocation(unsigned bool, baseAddress, tileNum uint16) uint16 {
	if !unsigned {
		tileNum = uint16(int(int8(tileNum)) + ObjCount)
	}
	return baseAddress + (tileNum * CharCodeSize)
}

func (gpu *GPU) fetchCharCodeBytes(baseAddress, tileOffset uint16) (byte, byte) {
	charCodeAddress := baseAddress + (uint16(tileOffset) << 1)
	low := gpu.cpu.memory.get(charCodeAddress)
	high := gpu.cpu.memory.get(charCodeAddress + 1)
	return low, high
}

func (gpu *GPU) applyBGPalette(colour byte) (byte, byte, byte) {
	customColour := (gpu.cpu.getBGP() >> (colour * ColourSize)) & ColourMask
	return gpu.applyCustomPalette(customColour)
}

func (gpu *GPU) fetchBitPair(xPos, low, high byte) byte {
	bitOffset := xPos & CharCodeMask
	bitL := (low >> (7 - bitOffset)) & 0x1
	bitH := (high >> (7 - bitOffset)) & 0x1
	return (bitH << 1) | bitL
}

func (gpu *GPU) getSpritePixel(e *oamEntry, x, y byte) (byte, byte, byte, bool) {
	tileX := byte(int16(x) - e.x)
	tileY := byte(int16(y) - e.y)

	if e.xFlip() {
		tileX = 7 - tileX
	}
	if e.yFlip() {
		tileY = e.height - 1 - tileY
	}
	tileNum := e.tileNum
	if e.height == 16 {
		tileNum &^= 0x01
		if tileY >= 8 {
			tileNum++
		}
	}

	charCode := uint16(tileY & CharCodeMask)
	spriteAddress := gpu.getSpriteAddress(tileNum)
	low, high := gpu.fetchSpriteData(spriteAddress, charCode)
	colour := gpu.fetchBitPair(tileX, low, high)
	if colour == 0 {
		return 0, 0, 0, false
	}

	palettedPixel := gpu.applySpritePalette(e, colour)
	r, g, b := gpu.applyCustomPalette(palettedPixel)
	return r, g, b, true
}

func (gpu *GPU) getSpriteAddress(tileNum byte) uint16 {
	return SpriteStartAddress + (uint16(tileNum) * SpriteDataSize)
}

func (gpu *GPU) fetchSpriteData(spriteAddress, charCode uint16) (byte, byte) {
	low := gpu.cpu.memory.get(spriteAddress + (charCode << 1))
	high := gpu.cpu.memory.get(spriteAddress + (charCode << 1) + 1)
	return low, high
}

func (gpu *GPU) applySpritePalette(e *oamEntry, colour byte) byte {
	palReg := gpu.cpu.getOBP0()
	if e.palSelector() {
		palReg = gpu.cpu.getOBP1()
	}

	return (palReg >> uint((colour * ColourSize))) & ColourMask
}

type oamEntry struct {
	y         int16
	x         int16
	height    byte
	tileNum   byte
	flagsByte byte
}

func (e *oamEntry) behindBG() bool    { return e.flagsByte&0x80 != 0 }
func (e *oamEntry) yFlip() bool       { return e.flagsByte&0x40 != 0 }
func (e *oamEntry) xFlip() bool       { return e.flagsByte&0x20 != 0 }
func (e *oamEntry) palSelector() bool { return e.flagsByte&0x10 != 0 }

func yInSprite(y byte, spriteY int16, height int) bool {
	return int16(y) >= spriteY && int16(y) < spriteY+int16(height)
}

// func (gpu *GPU) ParseSprites() {
// 	gpu.parseOAMForScanline(gpu.cpu.getLY())
// }

func (gpu *GPU) parseOAMForScanline(scanline byte) {
	height := 8

	gpu.oams = gpu.oams[:0]
	// search all sprites, limit total found to 10 per scanline
	for i := 0; len(gpu.oams) < 10 && i < 40; i++ {
		addr := 0xFE00 + uint16(i*4)
		// spriteY := int16(gpu.cpu.memory.get(addr)) - 16
		spriteY := int16(gpu.cpu.memory.sram[addr-0xFE00]) - 16
		if yInSprite(scanline, spriteY, height) {
			gpu.oams = append(gpu.oams, &oamEntry{
				y: spriteY,
				// x:         int16(gpu.cpu.memory.get(addr+1)) - 8,
				x:      int16(gpu.cpu.memory.sram[addr-0xFE00+1]) - 8,
				height: byte(height),
				// tileNum:   gpu.cpu.memory.get(addr + 2),
				tileNum: gpu.cpu.memory.sram[addr-0xFE00+2],
				// flagsByte: gpu.cpu.memory.get(addr + 3),
				flagsByte: gpu.cpu.memory.sram[addr-0xFE00+3],
			})
		}
	}

	sort.Stable(sortableOAM(gpu.oams))
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

func (status Status) mode() Mode {
	return Mode(status & ModeMask)
}

func (status Status) setMode(mode Mode) Status {
	return Status((byte(status) & StatusModeResetMask) | byte(mode))
}

func (status Status) isModeInterruptSet() bool {
	mode := status.mode()
	if mode == TransferringMode {
		return false
	}
	return utils.IsSet(byte(mode)+3, byte(status))
}

func (status Status) setMatchFlag() Status {
	return Status(utils.SetBit(byte(MatchFlag), byte(status), 1))
}

func (status Status) resetMatchFlag() Status {
	return Status(utils.SetBit(byte(MatchFlag), byte(status), 0))
}

func (status Status) isMatchInterruptSet() bool {
	return utils.IsSet(byte(MatchInterrupt), byte(status))
}

func (gpu *GPU) getStatus() Status {
	return Status(gpu.cpu.getSTAT())
}

func (gpu *GPU) setStatus(status Status) {
	gpu.cpu.setSTAT(byte(status))
}

func (gpu *GPU) setMode(mode Mode) {
	gpu.setStatus(gpu.getStatus().setMode(mode))
}
