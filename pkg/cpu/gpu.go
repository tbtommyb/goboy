package cpu

import (
	"sort"

	"github.com/tbtommyb/goboy/pkg/constants"
	"github.com/tbtommyb/goboy/pkg/utils"
)

const (
	MaxScanline                byte = 153
	MaxVisibleScanline              = 144
	ModeMask                        = 3
	ScrollXOffset                   = 7
	CharCodeMask                    = 7
	CharCodeShift                   = 1
	CharCodeSize                    = 16
	TilePixelSize                   = 8
	TileRowSize                     = 32
	ColourSize                      = 2
	ColourMask                      = 3
	SpriteDataSize                  = 16
	SpritePixelSize                 = 8
	ObjCount                        = 128
	SearchingOAMModeCycleBound      = 376
	TransferringModeCycleBound      = 302
	StatusModeResetMask             = 0xFC
	CyclesPerScanline          int  = 456
)

type Mode byte

const (
	AccessEnabledMode Mode = 0
	VBlankMode             = 1
	SearchingOAMMode       = 2
	TransferringMode       = 3
)

type GPU struct {
	mode          Mode
	cpu           *CPU
	display       DisplayInterface
	cyclesCounter int
	oams          []*oamEntry
}

type DisplayInterface interface {
	WritePixel(x, y, r, g, b, a byte)
}

func InitGPU(cpu *CPU) *GPU {
	return &GPU{
		mode: SearchingOAMMode,
		cpu:  cpu,
	}
}

var standardPalette = [][]byte{
	{0xff, 0xff, 0xff},
	{0xaa, 0xaa, 0xaa},
	{0x55, 0x55, 0x55},
	{0x00, 0x00, 0x00},
}

func (gpu *GPU) update(cycles uint) {
	gpu.setLCDStatus(gpu.cyclesCounter)
	if !gpu.cpu.isLCDCSet(LCDDisplayEnable) {
		gpu.cyclesCounter = CyclesPerScanline
		return
	}

	gpu.cyclesCounter -= int(cycles)

	if gpu.cyclesCounter > 0 {
		return
	}

	scanline := gpu.incrementScanline()
	gpu.cyclesCounter = CyclesPerScanline

	if scanline == MaxVisibleScanline {
		gpu.requestInterrupt(0)
	} else if scanline > MaxScanline {
		gpu.resetScanline()
	} else if scanline < MaxVisibleScanline {
		gpu.renderLine()
	}
}

func (gpu *GPU) resetScanline() {
	gpu.cpu.setLY(0)
}

func (gpu *GPU) incrementScanline() byte {
	currentScanline := gpu.cpu.getLY()
	currentScanline++
	if currentScanline > MaxScanline {
		currentScanline = 0
	}
	gpu.cpu.setLY(currentScanline)
	return currentScanline
}

func (gpu *GPU) renderLine() {
	scanline := gpu.cpu.getLY()
	if gpu.cpu.isLCDCSet(WindowDisplayPriority) {
		gpu.renderTiles()
	}
	if gpu.cpu.isLCDCSet(SpriteEnable) {
		gpu.renderSprites(gpu.oams, scanline)
	}
}

func (gpu *GPU) requestInterrupt(interrupt byte) {
	// TODO: why this here?
	gpu.parseOAMForScanline(gpu.cpu.getLY())
	gpu.cpu.requestInterrupt(interrupt)
}

func (gpu *GPU) bgTileDataAddress() (uint16, bool) {
	if gpu.cpu.isLCDCSet(DataSelect) {
		return 0x8000, true
	} else {
		return 0x8800, false
	}
}

func (gpu *GPU) bgTileMapStartAddress(usingWindow bool) uint16 {
	flag := LCDCFlag(BGTileMapDisplaySelect)
	if usingWindow {
		flag = WindowTileMapDisplaySelect
	}
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

func (gpu *GPU) renderTiles() {
	scrollX := gpu.cpu.getScrollX()
	scrollY := gpu.cpu.getScrollY()
	windowX := gpu.cpu.getWindowX() - ScrollXOffset
	windowY := gpu.cpu.getWindowY()
	scanline := gpu.cpu.getLY()

	var usingWindow bool
	if gpu.cpu.isLCDCSet(WindowDisplayEnable) {
		if windowY <= scanline {
			usingWindow = true
		}
	}

	tileDataAddress, unsig := gpu.bgTileDataAddress()
	startAddress := gpu.bgTileMapStartAddress(usingWindow)

	yPos := byte(scrollY + scanline)
	if usingWindow {
		yPos = scanline - windowY
	}

	for pixel := byte(0); pixel < byte(constants.ScreenWidth); pixel++ {
		xPos := byte(scrollX + pixel)
		if usingWindow && pixel >= windowX {
			xPos = pixel - windowX
		}

		tileNum := gpu.getTileNum(startAddress, xPos, yPos)
		tileLocation := getTileLocation(unsig, tileDataAddress, tileNum)

		charCode := yPos & CharCodeMask
		low, high := gpu.fetchCharCodeBytes(tileLocation, uint16(charCode))
		colour := gpu.fetchBitPair(xPos, low, high)
		r, g, b := gpu.applyBGPalette(colour)

		gpu.display.WritePixel(pixel, scanline, r, g, b, 0xff)
	}
}

func getTileLocation(unsigned bool, baseAddress, tileNum uint16) uint16 {
	if unsigned {
		return baseAddress + (tileNum * CharCodeSize)
	} else {
		return baseAddress + uint16((int(tileNum)+ObjCount)*CharCodeSize)
	}
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

func (gpu *GPU) renderSprites(oams []*oamEntry, scanline byte) {
	for _, e := range oams {
		startX := byte(0)
		if e.x > 0 {
			startX = byte(e.x)
		}
		endX := byte(e.x + SpritePixelSize)

		for x := startX; x < endX && x < byte(constants.ScreenWidth); x++ {
			// TODO: hide sprites
			if r, g, b, a := gpu.getSpritePixel(e, x, scanline); a {
				gpu.display.WritePixel(x, scanline, r, g, b, 0xff)
			}
		}
	}
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

func (gpu *GPU) ParseSprites() {
	gpu.parseOAMForScanline(gpu.cpu.getLY())
}

func (gpu *GPU) parseOAMForScanline(scanline byte) {
	height := 8

	gpu.oams = gpu.oams[:0]
	// search all sprites, limit total found to 10 per scanline
	for i := 0; len(gpu.oams) < 10 && i < 40; i++ {
		addr := 0xFE00 + uint16(i*4)
		spriteY := int16(gpu.cpu.memory.get(addr)) - 16
		if yInSprite(scanline, spriteY, height) {
			gpu.oams = append(gpu.oams, &oamEntry{
				y:         spriteY,
				x:         int16(gpu.cpu.memory.get(addr+1)) - 8,
				height:    byte(height),
				tileNum:   gpu.cpu.memory.get(addr + 2),
				flagsByte: gpu.cpu.memory.get(addr + 3),
			})
		}
	}

	sort.Stable(sortableOAM(gpu.oams))
}

type sortableOAM []*oamEntry

func (s sortableOAM) Less(i, j int) bool { return s[i].x < s[j].x }
func (s sortableOAM) Len() int           { return len(s) }
func (s sortableOAM) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (gpu *GPU) setLCDStatus(scanlineCounter int) {
	status := gpu.cpu.getSTAT()
	if !gpu.cpu.isLCDCSet(LCDDisplayEnable) {
		gpu.cpu.setLY(0)
		// status &= 252 // TODO improve this
		status = setStatusMode(status, VBlankMode)
		gpu.cpu.setSTAT(status)
		return
	}

	currentLine := gpu.cpu.getLY()
	currentMode := Mode(status & ModeMask)

	var newMode Mode
	if currentLine >= MaxVisibleScanline {
		newMode = VBlankMode
	} else if scanlineCounter >= SearchingOAMModeCycleBound {
		newMode = SearchingOAMMode
		if scanlineCounter == SearchingOAMModeCycleBound {
			gpu.parseOAMForScanline(currentLine)
		}
	} else if scanlineCounter >= TransferringModeCycleBound {
		newMode = TransferringMode
	} else {
		newMode = AccessEnabledMode
	}

	status = setStatusMode(status, newMode)
	requestInterrupt := modeInterruptSet(status, newMode)
	if requestInterrupt && (newMode != currentMode) {
		gpu.cpu.requestInterrupt(1)
	}

	if gpu.cpu.getLY() == gpu.cpu.getLYC() {
		status = utils.SetBit(byte(MatchFlag), status, 1)
		if utils.IsSet(byte(MatchInterrupt), status) {
			gpu.cpu.requestInterrupt(1)

		}
	} else {
		status = utils.SetBit(byte(MatchFlag), status, 0)
	}
	gpu.cpu.setSTAT(status)
}

func setStatusMode(statusRegister byte, mode Mode) byte {
	return (statusRegister & StatusModeResetMask) | byte(mode)
}

func modeInterruptSet(statusRegister byte, mode Mode) bool {
	if mode == TransferringMode {
		return false
	} else {
		return utils.IsSet(byte(mode)+3, statusRegister)
	}
}
