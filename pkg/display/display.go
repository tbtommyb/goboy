package display

import (
	"image"
)

type Display struct {
	buffer          *image.RGBA
	scanlineCounter int
	sysInterface    DisplayInterface
}

type DisplayInterface interface {
	GetScrollX() byte
	GetScrollY() byte
	TileColour(x, y byte) uint32
	GetLY() byte
	IncrementLY()
}

func (d *Display) renderLine(ly uint8) {
	tileY := ly + d.sysInterface.GetScrollY()

	for lcdX := uint8(0); lcdX < 160; lcdX++ {
		tileX := lcdX + d.sysInterface.GetScrollX()
		color := d.sysInterface.TileColour(tileX, tileY)
		pixel := int(ly)*int(160) + int(lcdX)

		d.buffer.Pix[4*pixel] = uint8((color >> 24) & 0xFF)
		d.buffer.Pix[4*pixel+1] = uint8((color >> 16) & 0xFF)
		d.buffer.Pix[4*pixel+2] = uint8((color >> 8) & 0xFF)
		d.buffer.Pix[4*pixel+3] = uint8(color & 0xFF)
	}
}

func (d *Display) Update(cycles int) {
	d.scanlineCounter += cycles

	if d.scanlineCounter < 456 {
		return
	}

	ly := d.sysInterface.GetLY()
	if ly < 144 { // Can only render the first 144 rows - the rest are never rendered
		d.renderLine(ly)
	}

	// Scanline ended here!
	d.scanlineCounter -= 456 // Save the extra cycles
	d.sysInterface.IncrementLY()
}

func (d *Display) Pixels() []uint8 {
	return d.buffer.Pix
}

func InitDisplay(di DisplayInterface) *Display {
	return &Display{
		buffer:       image.NewRGBA(image.Rect(0, 0, int(160), int(144))),
		sysInterface: di,
	}
}
