package display

import (
	"image"

	"github.com/tbtommyb/goboy/pkg/constants"
)

type Display struct {
	Buffer          *image.RGBA
	sysInterface    DisplayInterface
	scanlineCounter int
}

// TODO: change LY to scanline and keep LY as GB implementation detail?
type DisplayInterface interface {
	GetScrollX() byte
	GetScrollY() byte
	// TileColour(x, y byte) uint32
	GetScanline() byte
	IncrementScanline() byte
	ResetScanline()
	UpdateStatus(byte)
	DisplayEnabled() bool
	SetLCDStatus(scanlineCounter int)
	RenderLine()
}

// func (d *Display) renderLine(ly uint8) {
// 	tileY := ly + d.sysInterface.GetScrollY()

// 	for lcdX := uint8(0); lcdX < 160; lcdX++ {
// 		tileX := lcdX + d.sysInterface.GetScrollX()
// 		color := d.sysInterface.TileColour(tileX, tileY)
// 		pixel := int(ly)*int(160) + int(lcdX)

// 		d.buffer.Pix[4*pixel] = uint8((color >> 24) & 0xFF)
// 		d.buffer.Pix[4*pixel+1] = uint8((color >> 16) & 0xFF)
// 		d.buffer.Pix[4*pixel+2] = uint8((color >> 8) & 0xFF)
// 		d.buffer.Pix[4*pixel+3] = uint8(color & 0xFF)
// 	}
// }

func (d *Display) Update(cycles uint) {
	d.sysInterface.SetLCDStatus(d.scanlineCounter)
	// TODO: set LCD status
	if !d.sysInterface.DisplayEnabled() {
		d.scanlineCounter = 456
		return
	}

	d.scanlineCounter -= int(cycles)

	if d.scanlineCounter > 0 {
		return
	}

	scanline := d.sysInterface.IncrementScanline()
	d.scanlineCounter = 456

	if scanline == 144 {
		// TODO: trigger VBLANK
	} else if scanline > 153 {
		d.sysInterface.ResetScanline()
	} else if scanline < 144 {
		d.sysInterface.RenderLine()
	}
}

func (d *Display) Pixels() []uint8 {
	return d.Buffer.Pix
}

func InitDisplay(di DisplayInterface) *Display {
	return &Display{
		Buffer:       image.NewRGBA(image.Rect(0, 0, constants.ScreenWidth, constants.ScreenHeight)),
		sysInterface: di,
	}
}
