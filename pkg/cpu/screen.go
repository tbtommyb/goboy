package cpu

import (
	"github.com/tbtommyb/goboy/pkg/display"
	"github.com/tbtommyb/goboy/pkg/utils"
)

const (
	MaxLY     byte = 153
	White          = 0
	LightGrey      = 1
	DarkGrey       = 2
	Black          = 3
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

func (gpu *GPU) applyCustomPalette(val byte) (byte, byte, byte) {
	outVal := standardPalette[3-val]
	return outVal[0], outVal[1], outVal[2]
}

func (cpu *CPU) incrementLY() byte {
	currentScanline := cpu.getLY()
	currentScanline++
	if currentScanline > MaxLY {
		currentScanline = 0
	}
	cpu.setLY(currentScanline)
	return currentScanline
}

// func (gpu *GPU) TileColour(x byte, y byte) uint32 {
// 	// 32 tiles per row. y>>3 (same as y/8) gets the row. x>>3 (x/8) gets the columns
// 	tileMapOffset := (uint16(x) >> 3) + (uint16(y)>>3)*32
// 	tileSelectionAddress := gpu.bgTileMapStartAddress() + uint16(tileMapOffset)
// 	tileNumber := gpu.cpu.memory.get(tileSelectionAddress) // Which one of 256 tiles are to be shown
// 	tileDataAddress := gpu.bgTileDataAddress(tileNumber)   // Where the 16-bytes of the tile begin

// 	tileYOffset := (y & 0x7) * 2 // Each row in the tile takes 2 bytes
// 	tileXOffset := (x & 0x7)     // Each col in the tile is 1 bit
// 	pixelByte := tileDataAddress + uint16(tileYOffset)
// 	pixLow := (gpu.cpu.memory.get(pixelByte+1) >> (7 - tileXOffset)) & 0x1
// 	pixHigh := (gpu.cpu.memory.get(pixelByte) >> (7 - tileXOffset)) & 0x1
// 	colorNumber := (pixHigh << 1) | pixLow
// 	return ColourMap[colorNumber]
// }

func (gpu *GPU) bgTileDataAddress(tileNumber uint8) uint16 {
	var tileAddress uint16
	if gpu.cpu.isLCDCSet(DataSelect) {
		tileAddress = 0x8000
	} else {
		tileAddress = 0x8800
	}
	return tileAddress + uint16(tileNumber)*16
}

func (gpu *GPU) bgTileMapStartAddress() uint16 {
	if gpu.cpu.isLCDCSet(BGTileMapDisplaySelect) {
		return 0x9C00
	}
	return 0x9800
}

func (gpu *GPU) IncrementScanline() byte {
	return gpu.cpu.incrementLY()
}

func (gpu *GPU) GetScanline() byte {
	return gpu.cpu.getLY()
}

func (gpu *GPU) ResetScanline() {
	gpu.cpu.setLY(0)
}

func (gpu *GPU) UpdateStatus(status byte) {
	gpu.cpu.setSTAT(status)
}

func (gpu *GPU) GetScrollX() byte {
	return gpu.cpu.getScrollX()
}

func (gpu *GPU) GetScrollY() byte {
	return gpu.cpu.getScrollY()
}

func (gpu *GPU) DisplayEnabled() bool {
	return gpu.cpu.isLCDCSet(LCDDisplayEnable)
}

func (gpu *GPU) RenderLine() {
	if gpu.cpu.isLCDCSet(WindowDisplayPriority) {
		gpu.renderTiles()
	}
	if gpu.cpu.isLCDCSet(SpriteEnable) {
		// gpu.renderSprites()
	}
}

func (gpu *GPU) renderTiles() {
	var tileData, backgroundMemory uint16
	unsig := true

	scrollX := gpu.GetScrollX()
	// fmt.Printf("%x\n", scrollX)
	scrollY := gpu.GetScrollY()
	windowX := gpu.cpu.GetWindowX() - 7 // TODO: explain
	windowY := gpu.cpu.GetWindowY()

	usingWindow := false

	if gpu.cpu.isLCDCSet(WindowDisplayEnable) {
		if windowY <= gpu.cpu.getLY() {
			usingWindow = true
		}
	}

	// TILEDATA
	if gpu.cpu.isLCDCSet(DataSelect) {
		tileData = 0x8000
	} else {
		tileData = 0x8800
		unsig = false
	}

	// // BACKGROUND MEM
	if usingWindow {
		if gpu.cpu.isLCDCSet(WindowTileMapDisplaySelect) {
			backgroundMemory = 0x9C00
		} else {
			backgroundMemory = 0x9800
		}
	} else {
		if gpu.cpu.isLCDCSet(BGTileMapDisplaySelect) {
			backgroundMemory = 0x9C00
		} else {
			backgroundMemory = 0x9800
		}
	}

	var yPos byte
	if !usingWindow {
		yPos = scrollY + gpu.cpu.getLY()
	} else {
		yPos = gpu.cpu.getLY() - windowY
	}
	// tileRow := uint16(byte(yPos/8) * 32)

	for pixel := byte(0); pixel < 160; pixel++ {
		xPos := byte(pixel + scrollX)
		if usingWindow {
			if pixel >= windowX {
				xPos = pixel - windowX
			}
		}
		tileNumY, tileNumX := uint16(yPos>>3), uint16(xPos>>3)
		// realColour := gpu.TileColour(xPos, yPos)

		// tileCol := uint16(xPos / 8)

		tileAddress := uint16(backgroundMemory + tileNumY*32 + tileNumX)
		tileNum := uint16(gpu.cpu.memory.get(tileAddress))
		tileLocation := tileData
		if unsig {
			tileLocation += tileNum * 16
		} else {
			tileLocation += uint16((int(tileNum) + 128) * 16)
		}

		mapBitY, mapBitX := yPos&0x07, xPos&0x07

		dataByteL := gpu.cpu.memory.get(tileLocation + (uint16(mapBitY) << 1))
		dataByteH := gpu.cpu.memory.get(tileLocation + (uint16(mapBitY) << 1) + 1)
		dataBitL := (dataByteL >> (7 - mapBitX)) & 0x1
		dataBitH := (dataByteH >> (7 - mapBitX)) & 0x1
		colourBit := (dataBitH << 1) | dataBitL

		palettedPixel := (gpu.cpu.getGBP() >> (colourBit * 2)) & 0x03
		r, g, b := gpu.applyCustomPalette(palettedPixel)

		yIdx := int(gpu.cpu.getLY())*int(160) + int(pixel)
		gpu.display.Buffer.Pix[4*yIdx] = byte(r)
		gpu.display.Buffer.Pix[4*yIdx+1] = byte(g)
		gpu.display.Buffer.Pix[4*yIdx+2] = byte(b)
		gpu.display.Buffer.Pix[4*yIdx+3] = 0xff
	}
}

func (gpu *GPU) getColour(colourNum byte, address uint16) byte {
	colour := byte(White)
	palette := gpu.cpu.memory.get(address)
	var hi, lo byte

	switch colourNum {
	case 0:
		hi = 1
		lo = 0
	case 1:
		hi = 3
		lo = 2
	case 2:
		hi = 5
		lo = 4
	case 3:
		hi = 7
		lo = 6
	}
	if utils.IsSet(hi, palette) {
		colour <<= 1
	}
	if utils.IsSet(lo, palette) {
		colour |= 1
	}

	return colour
}

func (gpu *GPU) SetLCDStatus(scanlineCounter int) {
	status := gpu.cpu.getSTAT()
	if !gpu.cpu.isLCDCSet(LCDDisplayEnable) {
		gpu.ResetScanline()
		status &= 252 // TODO improve this
		status = utils.SetBit(0, status, 1)
		gpu.UpdateStatus(status)
		return
	}

	currentLine := gpu.GetScanline()
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
		// request interrupt
	}

	if gpu.cpu.getLY() == gpu.cpu.getLYC() {
		status = utils.SetBit(2, status, 1)
		if utils.IsSet(6, status) {
			// request interrupt

		}
	} else {
		status = utils.SetBit(2, status, 0)
	}
	gpu.UpdateStatus(status)
}
