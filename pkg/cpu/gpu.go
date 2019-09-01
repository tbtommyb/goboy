package cpu

import (
	"sort"

	"github.com/tbtommyb/goboy/pkg/display"
	"github.com/tbtommyb/goboy/pkg/utils"
)

const (
	MaxLY byte = 153
)

type GPU struct {
	mode    byte
	cpu     *CPU
	display *display.Display
	oams    []*oamEntry
}

func InitGPU(cpu *CPU) *GPU {
	return &GPU{
		mode: 2,
		cpu:  cpu,
	}
}

var standardPalette = [][]byte{
	{0x00, 0x00, 0x00},
	{0x55, 0x55, 0x55},
	{0xaa, 0xaa, 0xaa},
	{0xff, 0xff, 0xff},
}

func (gpu *GPU) DisplayEnabled() bool {
	return gpu.cpu.isLCDCSet(LCDDisplayEnable)
}

func (gpu *GPU) IncrementScanline() byte {
	currentScanline := gpu.cpu.getLY()
	currentScanline++
	if currentScanline > MaxLY {
		currentScanline = 0
	}
	gpu.cpu.setLY(currentScanline)
	return currentScanline
}

func (gpu *GPU) ResetScanline() {
	gpu.cpu.setLY(0)
}

func (gpu *GPU) RenderLine() {
	scanline := gpu.cpu.getLY()
	if gpu.cpu.isLCDCSet(WindowDisplayPriority) {
		gpu.renderTiles()
	}
	if gpu.cpu.isLCDCSet(SpriteEnable) {
		gpu.renderSprites(gpu.oams, scanline)
	}
}

func (gpu *GPU) RequestInterrupt(interrupt byte) {
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

func (gpu *GPU) bgTileMapStartAddress(flag LCDCFlag) uint16 {
	if gpu.cpu.isLCDCSet(BGTileMapDisplaySelect) {
		return 0x9C00
	}
	return 0x9800
}

func (gpu *GPU) applyCustomPalette(val byte) (byte, byte, byte) {
	outVal := standardPalette[3-val]
	return outVal[0], outVal[1], outVal[2]
}

func (gpu *GPU) renderTiles() {
	scrollX := gpu.cpu.getScrollX()
	scrollY := gpu.cpu.getScrollY()
	windowX := gpu.cpu.getWindowX() - 7 // TODO: explain
	windowY := gpu.cpu.getWindowY()

	var usingWindow bool
	if gpu.cpu.isLCDCSet(WindowDisplayEnable) {
		if windowY <= gpu.cpu.getLY() {
			usingWindow = true
		}
	}

	tileDataAddress, unsig := gpu.bgTileDataAddress()

	var backgroundMemory uint16
	if usingWindow {
		backgroundMemory = gpu.bgTileMapStartAddress(WindowTileMapDisplaySelect)
	} else {
		backgroundMemory = gpu.bgTileMapStartAddress(BGTileMapDisplaySelect)
	}

	var yPos byte
	if usingWindow {
		yPos = gpu.cpu.getLY() - windowY
	} else {
		yPos = scrollY + gpu.cpu.getLY()
	}

	for pixel := byte(0); pixel < 160; pixel++ {
		xPos := byte(pixel + scrollX)
		if usingWindow {
			if pixel >= windowX {
				xPos = pixel - windowX
			}
		}
		tileNumY, tileNumX := uint16(yPos>>3), uint16(xPos>>3)

		tileAddress := uint16(backgroundMemory + tileNumY*32 + tileNumX)
		tileNum := uint16(gpu.cpu.memory.get(tileAddress))

		var tileLocation uint16
		if unsig {
			tileLocation = tileDataAddress + tileNum*16
		} else {
			tileLocation = tileDataAddress + uint16((int(tileNum)+128)*16)
		}

		mapBitY, mapBitX := yPos&0x07, xPos&0x07

		dataByteL := gpu.cpu.memory.get(tileLocation + (uint16(mapBitY) << 1))
		dataByteH := gpu.cpu.memory.get(tileLocation + (uint16(mapBitY) << 1) + 1)
		dataBitL := (dataByteL >> (7 - mapBitX)) & 0x1
		dataBitH := (dataByteH >> (7 - mapBitX)) & 0x1
		colourBit := (dataBitH << 1) | dataBitL

		palettedPixel := (gpu.cpu.getBGP() >> (colourBit * 2)) & 0x03
		r, g, b := gpu.applyCustomPalette(palettedPixel)

		yIdx := int(gpu.cpu.getLY())*int(160) + int(pixel)
		gpu.display.Buffer.Pix[4*yIdx] = byte(r)
		gpu.display.Buffer.Pix[4*yIdx+1] = byte(g)
		gpu.display.Buffer.Pix[4*yIdx+2] = byte(b)
		gpu.display.Buffer.Pix[4*yIdx+3] = 0xff
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
	mapBitY, mapBitX := tileY&0x07, tileX&0x07

	dataByteL := gpu.cpu.memory.get(0x8000 + (uint16(mapBitY) << 1))
	dataByteH := gpu.cpu.memory.get(0x8000 + (uint16(mapBitY) << 1) + 1)
	dataBitL := (dataByteL >> (7 - mapBitX)) & 0x1
	dataBitH := (dataByteH >> (7 - mapBitX)) & 0x1
	colourBit := (dataBitH << 1) | dataBitL

	if colourBit == 0 {
		return 0, 0, 0, false
	}
	palReg := gpu.cpu.getOBP0()
	if utils.IsSet(4, e.flagsByte) {
		palReg = gpu.cpu.getOBP1()
	}

	palettedPixel := (palReg >> uint((colourBit * 2))) & 0x03
	r, g, b := gpu.applyCustomPalette(palettedPixel)
	return r, g, b, true
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
		// fmt.Printf("addr: %x, val: %x\n", addr, gpu.cpu.memory.get(addr))
		// fmt.Printf("%x, %x, %x\n", scanline, spriteY, height)
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

func (gpu *GPU) renderSprites(oams []*oamEntry, scanline byte) {
	for _, e := range oams {
		startX := byte(0)
		if e.x > 0 {
			startX = byte(e.x)
		}
		endX := byte(e.x + 8)

		for x := startX; x < endX && x < 160; x++ {
			// TODO: hide sprite?
			if r, g, b, a := gpu.getSpritePixel(e, byte(x), byte(scanline)); a {
				gpu.setFramebufferPixel(x, scanline, r, g, b)
			}
		}
	}
}

func (gpu *GPU) setFramebufferPixel(xByte, yByte, r, g, b byte) {
	x, y := int(xByte), int(yByte)
	yIdx := y * 160 * 4
	gpu.display.Buffer.Pix[yIdx+x*4] = byte(r)
	gpu.display.Buffer.Pix[yIdx+x*4+1] = byte(g)
	gpu.display.Buffer.Pix[yIdx+x*4+2] = byte(b)
	gpu.display.Buffer.Pix[yIdx+x*4+3] = 0xff
}

func (gpu *GPU) SetLCDStatus(scanlineCounter int) {
	status := gpu.cpu.getSTAT()
	if !gpu.cpu.isLCDCSet(LCDDisplayEnable) {
		gpu.ResetScanline()
		status &= 252 // TODO improve this
		status = utils.SetBit(0, status, 1)
		gpu.cpu.setSTAT(status)
		return
	}

	currentLine := gpu.cpu.getLY()
	currentMode := status & 0x3
	mode := byte(0)
	requestInterrupt := false

	if currentLine >= 144 {
		// In VBLANK
		mode = 1
		status = utils.SetBit(0, status, 1)
		status = utils.SetBit(1, status, 0)
		requestInterrupt = utils.IsSet(4, status)
	} else {
		mode2bounds := 456 - 80
		mode3bounds := mode2bounds - 72

		if scanlineCounter >= mode2bounds {
			mode = 2
			status = utils.SetBit(1, status, 1)
			status = utils.SetBit(0, status, 0)
			requestInterrupt = utils.IsSet(5, status)
			if scanlineCounter == mode2bounds {
				gpu.parseOAMForScanline(currentLine)
			}
		} else if scanlineCounter >= mode3bounds {
			mode = 3
			status = utils.SetBit(1, status, 1)
			status = utils.SetBit(0, status, 1)
		} else {
			mode = 0
			status = utils.SetBit(1, status, 0)
			status = utils.SetBit(0, status, 0)
			requestInterrupt = utils.IsSet(3, status)
		}
	}

	if requestInterrupt && (mode != currentMode) {
		gpu.cpu.requestInterrupt(1)
	}

	if gpu.cpu.getLY() == gpu.cpu.getLYC() {
		status = utils.SetBit(2, status, 1)
		if utils.IsSet(6, status) {
			gpu.cpu.requestInterrupt(1)

		}
	} else {
		status = utils.SetBit(2, status, 0)
	}
	gpu.cpu.setSTAT(status)
}
