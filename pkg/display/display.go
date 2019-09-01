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

type DisplayInterface interface {
	SetLCDStatus(scanlineCounter int)
	DisplayEnabled() bool
	IncrementScanline() byte
	ResetScanline()
	RenderLine()
	RequestInterrupt(byte)
	ParseSprites()
}

func (d *Display) Update(cycles uint) {
	d.sysInterface.SetLCDStatus(d.scanlineCounter)
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
		d.sysInterface.RequestInterrupt(0)
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
