package cpu

import (
	"image"
)

type Display struct {
	internalImage   *image.RGBA
	scanlineCounter int
	cpu             *CPU
}

func (display *Display) drawBackgroundLine(ly uint8) {
	tileY := ly + display.cpu.memory.scrollY() // This will overflow as needed!

	for lcdX := uint8(0); lcdX < 160; lcdX++ {
		tileX := lcdX + display.cpu.memory.scrollX() // This will overflow as needed
		color := display.cpu.memory.backgroundPixelAt(tileX, tileY)
		pixel := int(ly)*int(160) + int(lcdX)

		display.internalImage.Pix[4*pixel] = uint8((color >> 24) & 0xFF)
		display.internalImage.Pix[4*pixel+1] = uint8((color >> 16) & 0xFF)
		display.internalImage.Pix[4*pixel+2] = uint8((color >> 8) & 0xFF)
		display.internalImage.Pix[4*pixel+3] = uint8(color & 0xFF)
	}
}
func (display *Display) renderLine(ly uint8) {
	display.drawBackgroundLine(ly)

}
func (display *Display) Update(cycles int) {
	display.scanlineCounter += cycles

	if display.scanlineCounter < 456 {
		return
	}

	ly := display.cpu.memory.getLY()
	if ly < 144 { // Can only render the first 144 rows - the rest are never rendered
		display.renderLine(ly)
	}

	// Scanline ended here!
	display.scanlineCounter -= 456 // Save the extra cycles
	display.cpu.memory.incrementLY()
}

func (d *Display) Pixels() []uint8 {
	return d.internalImage.Pix
}

func InitDisplay(cpu *CPU) *Display {
	return &Display{
		internalImage: image.NewRGBA(image.Rect(0, 0, int(160), int(144))),
		cpu:           cpu,
	}
}
