package cpu

import "image"

type Display struct {
	internalImage *image.RGBA
}

func (d *Display) Pixels() []uint8 {
	return d.internalImage.Pix
}

func InitDisplay() *Display {
	return &Display{
		internalImage: image.NewRGBA(image.Rect(0, 0, int(160), int(144))),
	}
}
