package display

import (
	"image"

	"github.com/tbtommyb/goboy/pkg/constants"
)

type Display struct {
	buffer          *image.RGBA
	scanlineCounter int
	sysInterface    DisplayInterface
}

// TODO: change LY to scanline and keep LY as GB implementation detail?
type DisplayInterface interface {
	GetScrollX() byte
	GetScrollY() byte
	TileColour(x, y byte) uint32
	GetScanline() byte
	IncrementScanline()
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

	ly := d.sysInterface.GetScanline()
	if ly < 144 { // Can only render the first 144 rows - the rest are never rendered
		d.renderLine(ly)
	}

	// Scanline ended here!
	d.scanlineCounter -= 456 // Save the extra cycles
	d.sysInterface.IncrementScanline()
}

func (d *Display) Pixels() []uint8 {
	return d.buffer.Pix
}

func InitDisplay(di DisplayInterface) *Display {
	return &Display{
		buffer:       image.NewRGBA(image.Rect(0, 0, constants.ScreenWidth, constants.ScreenHeight)),
		sysInterface: di,
	}
}
