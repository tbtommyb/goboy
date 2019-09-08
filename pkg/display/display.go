package display

import (
	"image"

	"github.com/tbtommyb/goboy/pkg/constants"
)

type Display struct {
	buffer *image.RGBA
}

func (d *Display) WritePixel(x, y, r, g, b, a byte) {
	yIdx := int(y)*160 + int(x)
	d.buffer.Pix[yIdx*4] = byte(r)
	d.buffer.Pix[yIdx*4+1] = byte(g)
	d.buffer.Pix[yIdx*4+2] = byte(b)
	d.buffer.Pix[yIdx*4+3] = 0xff
}

func (d *Display) Pixels() []uint8 {
	return d.buffer.Pix
}

func Init() *Display {
	return &Display{
		buffer: image.NewRGBA(image.Rect(0, 0, constants.ScreenWidth, constants.ScreenHeight)),
	}
}
