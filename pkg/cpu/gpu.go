package cpu

// TODO
// fix bug in rendering Mario
// improve handling of where modes are stored
import (
	"sort"

	"github.com/tbtommyb/goboy/pkg/constants"
	"github.com/tbtommyb/goboy/pkg/utils"
)

const (
	MaxScanline                byte = 153
	MaxVisibleScanline              = 144
	ScanlinesPerVBlank              = 10
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
	CyclesPerScanline          uint = 456
)

type Mode byte

const (
	HBlankMode       Mode = 0
	VBlankMode            = 1
	SearchingOAMMode      = 2
	TransferringMode      = 3
)

type GPU struct {
	mode          Mode
	cpu           *CPU
	display       DisplayInterface
	cyclesCounter uint
	oams          []*oamEntry
	// for oam sprite priority
	BGMask     [160]bool
	SpriteMask [160]bool
}

type DisplayInterface interface {
	WritePixel(x, y, r, g, b, a byte)
}

func InitGPU(cpu *CPU) *GPU {
	gpu := &GPU{
		mode: SearchingOAMMode,
		cpu:  cpu,
	}
	// TODO: tidy this up
	status := gpu.cpu.getSTAT()
	status = setStatusMode(status, SearchingOAMMode)
	gpu.cpu.setSTAT(status)
	return gpu
}

var standardPalette = [][]byte{
	{0xff, 0xff, 0xff},
	{0xaa, 0xaa, 0xaa},
	{0x55, 0x55, 0x55},
	{0x00, 0x00, 0x00},
}

func (gpu *GPU) update(_ uint) {
	status := gpu.cpu.getSTAT()
	if !gpu.cpu.isLCDCSet(LCDDisplayEnable) {
		gpu.resetScanline()
		// status &= 252 // TODO improve this
		status = setStatusMode(status, VBlankMode)
		gpu.cpu.setSTAT(status)
		return
	}
	currentLine := gpu.cpu.getLY()
	currentMode := Mode(status & ModeMask)
	newMode := currentMode

	gpu.cyclesCounter += 1
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
			gpu.renderLine(currentLine)
		}
	case HBlankMode:
		if gpu.cyclesCounter >= 204 {
			gpu.cyclesCounter = 0
			gpu.incrementScanline()

			if currentLine == MaxVisibleScanline-1 && currentMode != VBlankMode {
				newMode = VBlankMode
				gpu.requestInterrupt(0)
			} else {
				newMode = SearchingOAMMode
			}
		}
	case VBlankMode:
		if gpu.cyclesCounter >= CyclesPerScanline*4 {
			gpu.cyclesCounter = 0
			newScanline := gpu.incrementScanline()
			if newScanline > MaxScanline {
				newMode = SearchingOAMMode
				gpu.resetScanline()
			}
		}
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
	gpu.mode = newMode

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

func (gpu *GPU) renderLine(scanline byte) {
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

func (gpu *GPU) requestInterrupt(interrupt byte) {
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

func (gpu *GPU) renderWindow(scanline byte) {
	winY := scanline - gpu.cpu.getWindowY()
	winStartX := int(gpu.cpu.getWindowX()) - ScrollXOffset

	for x := winStartX; x < constants.ScreenWidth; x++ {
		if x < 0 {
			continue
		}

		pixel := gpu.getWindowPixel(byte(x-winStartX), winY)
		if pixel != 0 {
			gpu.BGMask[x] = true
		}
		r, g, b := gpu.applyBGPalette(pixel)
		gpu.display.WritePixel(byte(x), scanline, r, g, b, 0xff)
	}
}

func (gpu *GPU) getWindowPixel(x, y byte) byte {
	startAddress := gpu.windowTileMapStartAddress()
	dataAddress, unsig := gpu.bgTileDataAddress()
	tileNum := gpu.getTileNum(startAddress, x, y)
	tileLocation := getTileLocation(unsig, dataAddress, tileNum)
	charCode := y & CharCodeMask
	low, high := gpu.fetchCharCodeBytes(tileLocation, uint16(charCode))
	return gpu.fetchBitPair(x, low, high)
}

func (gpu *GPU) renderBackground(scanline byte) {
	scrollX := gpu.cpu.getScrollX()
	scrollY := gpu.cpu.getScrollY()

	tileDataAddress, unsig := gpu.bgTileDataAddress()
	startAddress := gpu.bgTileMapStartAddress()

	yPos := byte(scrollY + scanline)

	for pixel := byte(0); pixel < byte(constants.ScreenWidth); pixel++ {
		xPos := byte(scrollX + pixel)

		tileNum := gpu.getTileNum(startAddress, xPos, yPos)
		tileLocation := getTileLocation(unsig, tileDataAddress, tileNum)

		charCode := yPos & CharCodeMask
		low, high := gpu.fetchCharCodeBytes(tileLocation, uint16(charCode))
		colour := gpu.fetchBitPair(xPos, low, high)
		if colour != 0 {
			gpu.BGMask[pixel] = true
		}
		r, g, b := gpu.applyBGPalette(colour)

		gpu.display.WritePixel(pixel, scanline, r, g, b, 0xff)
	}
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

// func (gpu *GPU) setLCDStatus(scanlineCounter int) {
// 	status := gpu.cpu.getSTAT()
// 	if !gpu.cpu.isLCDCSet(LCDDisplayEnable) {
// 		gpu.cpu.setLY(0)
// 		// status &= 252 // TODO improve this
// 		status = setStatusMode(status, VBlankMode)
// 		gpu.cpu.setSTAT(status)
// 		return
// 	}

// 	currentLine := gpu.cpu.getLY()
// 	currentMode := Mode(status & ModeMask)

// 	var newMode Mode
// 	if currentLine >= MaxVisibleScanline {
// 		newMode = VBlankMode
// 	} else if scanlineCounter >= SearchingOAMModeCycleBound {
// 		newMode = SearchingOAMMode
// 		if scanlineCounter == SearchingOAMModeCycleBound {
// 			gpu.parseOAMForScanline(currentLine)
// 		}
// 	} else if scanlineCounter >= TransferringModeCycleBound {
// 		newMode = TransferringMode
// 	} else {
// 		newMode = AccessEnabledMode
// 	}

// 	status = setStatusMode(status, newMode)
// 	requestInterrupt := modeInterruptSet(status, newMode)
// 	if requestInterrupt && (newMode != currentMode) {
// 		gpu.cpu.requestInterrupt(1)
// 	}

// 	if gpu.cpu.getLY() == gpu.cpu.getLYC() {
// 		status = utils.SetBit(byte(MatchFlag), status, 1)
// 		if utils.IsSet(byte(MatchInterrupt), status) {
// 			gpu.cpu.requestInterrupt(1)

// 		}
// 	} else {
// 		status = utils.SetBit(byte(MatchFlag), status, 0)
// 	}
// 	gpu.cpu.setSTAT(status)
// }

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

// -----------
func (gpu *GPU) writeOAM(addr uint16, val byte) {
	if !(gpu.mode == SearchingOAMMode || gpu.mode == TransferringMode) {
		gpu.cpu.memory.sram[addr-0xFE00] = val
	}
}

func (gpu *GPU) readOAM(addr uint16) byte {
	if !(gpu.mode == SearchingOAMMode || gpu.mode == TransferringMode) {
		return gpu.cpu.memory.sram[addr-0xFE00]
	}
	return 0xff
}
