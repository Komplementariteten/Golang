package img

import (
	"errors"
	"image"
	"image/color"
	"math"
	"math/rand"
)

const RED_BIT_MASK = 0xFF000000
const RED_BIT_SHIFT = 030
const GREEN_BIT_MASK = 0x00FF0000
const GREEN_BIT_SHIFT = 020
const BLUE_BIT_MASK = 0x0000FF00
const BLUE_BIT_SHIFT = 010
const ALPHA_BIT_MASK = 0x000000FF
const ALPHABIT_SHIFT = 0

type RgbImage struct {
	Raw []uint32
	Bounds *image.Rectangle
	cache interface{}
}

func NewImage(width int, height int) (*RgbImage, error) {
	a0 := image.Point{Y:0, X:0}
	a1 := image.Point{X:width, Y:height}
	size := width * height
	bounds := &image.Rectangle{Min:a0, Max:a1}
	raw := make([]uint32, size)
	i := &RgbImage{Bounds:bounds, Raw:raw}
	return i, nil
}

func (img *RgbImage) Fill(color *color.RGBA) {
	colorValue := ColorToUint(color)
	for index ,_ := range img.Raw {
		img.Raw[index] = colorValue
	}
}

func (img *RgbImage) SetNoice(col *color.RGBA, distribution float64) error {
	if distribution >= 1 {
		return errors.New("distribution must be lower than 1")
	}

	items := int(math.Floor(float64(len(img.Raw)) * distribution))
	uintCol := ColorToUint(col)

	for i := 0; i < items; i++ {
		x := rand.Intn(img.Bounds.Dx())
		y := rand.Intn(img.Bounds.Dy())
		if x > 0 && y > 0 {
			index := img.Bounds.Dx() * y + x
			img.Raw[index] = uintCol
		}
	}
	return nil
}

func ColorToUint(color *color.RGBA) uint32 {
	red := uint32(color.R) << RED_BIT_SHIFT
	green := uint32(color.G) << GREEN_BIT_SHIFT
	blue := uint32(color.B) << BLUE_BIT_SHIFT
	alpha := uint32(color.A) << ALPHABIT_SHIFT
	return red + green + blue + alpha
}

func UintToColor(num uint32) *color.RGBA {
	red := uint8((num & RED_BIT_MASK) >> RED_BIT_SHIFT)
	green := uint8((num & GREEN_BIT_MASK) >> GREEN_BIT_SHIFT)
	blue := uint8((num & BLUE_BIT_MASK) >> BLUE_BIT_SHIFT)
	alpha := uint8((num & ALPHA_BIT_MASK) >> ALPHABIT_SHIFT)

	c := &color.RGBA{A:alpha, R:red, G:green, B:blue}
	return c
}


func (img *RgbImage) Paint() *image.RGBA {
	rgba := image.NewRGBA(*img.Bounds)
	img.cache = rgba
	rows := img.Bounds.Dx()
	lastColor := uint32(0)
	var col *color.RGBA
	position := image.Point{ X:0, Y:0 }
	for index, colorValue := range img.Raw {
		if index >= (rows * position.Y) {
			position.Y++
			position.X = 0
		}
		position.X++
		if lastColor != colorValue {
			col = UintToColor(colorValue)
			lastColor = colorValue
		}
		rgba.Set(position.X, position.Y, col)
	}
	img.cache = rgba
	return rgba
}