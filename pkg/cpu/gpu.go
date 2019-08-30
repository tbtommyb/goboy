package cpu

import (
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
	if gpu.cpu.isLCDCSet(WindowDisplayPriority) {
		gpu.renderTiles()
	}
	if gpu.cpu.isLCDCSet(SpriteEnable) {
		gpu.renderSprites()
	}
}

func (gpu *GPU) RequestInterrupt(interrupt byte) {
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

func (gpu *GPU) renderSprites() {
	use8x16 := gpu.cpu.isLCDCSet(SpriteSize)

	for sprite := 0; sprite < 40; sprite++ {
		index := byte(sprite * 4)
		// TODO: remove 0xFE00 (sprite memory base)
		yPos := gpu.cpu.memory.get(0xFE00+uint16(index)) - 16
		xPos := gpu.cpu.memory.get(0xFE00+uint16(index)+1) - 8
		tileLocation := gpu.cpu.memory.get(0xFE00 + uint16(index) + 2)
		attributes := gpu.cpu.memory.get(0xFE00 + uint16(index) + 3)

		yFlip := utils.IsSet(6, attributes)
		xFlip := utils.IsSet(5, attributes)

		scanline := int(gpu.cpu.getLY())

		ySize := int(8)
		if use8x16 {
			ySize = 16
		}

		if (scanline >= int(yPos)) && (scanline < (int(yPos) + ySize)) {
			line := int(scanline - int(yPos))

			if yFlip {
				line -= ySize
				line *= -1
			}

			line *= 2
			dataAddress := (0x8000 + uint16(tileLocation*16)) + uint16(line)
			data1 := gpu.cpu.memory.get(dataAddress)
			data2 := gpu.cpu.memory.get(dataAddress + 1)

			for tilePixel := int(7); tilePixel >= 0; tilePixel-- {
				colourBit := tilePixel
				if xFlip {
					colourBit -= 7
					colourBit *= -1
				}

				var colourNum int
				if utils.IsSet(byte(colourBit), data1) {
					colourNum = 1
				}
				colourNum <<= 1
				if utils.IsSet(byte(colourBit), data2) {
					colourNum |= 1
				}

				palReg := gpu.cpu.getOBP0()
				if utils.IsSet(4, attributes) {
					palReg = gpu.cpu.getOBP1()
				}

				palettedPixel := (palReg >> uint((colourBit * 2))) & 0x03
				r, g, b := gpu.applyCustomPalette(palettedPixel)

				xPix := 0 - tilePixel
				xPix += 7

				pixel := xPos + byte(xPix)

				yIdx := int(gpu.cpu.getLY())*int(160) + int(pixel)
				gpu.display.Buffer.Pix[4*yIdx] = byte(r)
				gpu.display.Buffer.Pix[4*yIdx+1] = byte(g)
				gpu.display.Buffer.Pix[4*yIdx+2] = byte(b)
				gpu.display.Buffer.Pix[4*yIdx+3] = 0xff

			}
		}

	}
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
