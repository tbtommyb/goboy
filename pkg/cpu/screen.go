package cpu

const (
	MaxLY byte = 153
)

var ColourMap = []uint32{0xFFFFFFFF, 0xB6B6B6FF, 0x676767FF, 0x000000FF}

func (cpu *CPU) incrementLY() {
	currentScanline := cpu.GetLY()
	currentScanline++
	if currentScanline > MaxLY {
		currentScanline = 0
	}
	cpu.setLY(currentScanline)
}

func (cpu *CPU) TileColour(x byte, y byte) uint32 {
	// 32 tiles per row. y>>3 (same as y/8) gets the row. x>>3 (x/8) gets the columns
	tileMapOffset := (uint16(x) >> 3) + (uint16(y)>>3)*32
	tileSelectionAddress := cpu.bgTileMapStartAddress() + uint16(tileMapOffset)
	tileNumber := cpu.memory.get(tileSelectionAddress)   // Which one of 256 tiles are to be shown
	tileDataAddress := cpu.bgTileDataAddress(tileNumber) // Where the 16-bytes of the tile begin

	tileYOffset := (y & 0x7) * 2 // Each row in the tile takes 2 bytes
	tileXOffset := (x & 0x7)     // Each col in the tile is 1 bit
	pixelByte := tileDataAddress + uint16(tileYOffset)
	pixLow := (cpu.memory.get(pixelByte+1) >> (7 - tileXOffset)) & 0x1
	pixHigh := (cpu.memory.get(pixelByte) >> (7 - tileXOffset)) & 0x1
	colorNumber := (pixHigh << 1) | pixLow
	return ColourMap[colorNumber]
}

func (cpu *CPU) bgTileDataAddress(tileNumber uint8) uint16 {
	var tileAddress uint16
	if cpu.isLCDCSet(DataSelect) {
		tileAddress = 0x8000
	} else {
		tileAddress = 0x8800
	}
	return tileAddress + uint16(tileNumber)*16
}

func (cpu *CPU) bgTileMapStartAddress() uint16 {
	if cpu.isLCDCSet(BGTileMapDisplaySelect) {
		return 0x9C00
	}
	return 0x9800
}

func (cpu *CPU) IncrementScanline() {
	cpu.incrementLY()
}

func (cpu *CPU) GetScanline() byte {
	return cpu.getLY()
}
