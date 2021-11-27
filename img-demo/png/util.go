package png

import (
	"image"
	"image/color"
)

type ColorFeature struct {
	Brightness         int
	Saturation         int
	RelativeBrightness int
	Chromaticity       int
}

type colorTestFnc func(map[color.Color]int, color.Color) color.Color
type colorHashFnc func(color.Color) int

func calculateChromaticity(color color.Color) int {
	r, g, b, a := color.RGBA()

}

func FindColorFeature(color color.Color) ColorFeature {

}

func FindSimilarColors(img image.Image, fnc colorTestFnc) map[color.Color]int {
	colorMap := make(map[color.Color]int)
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			color := fnc(colorMap, img.At(x, y))
			if color != nil {
				colorMap[color]++
			}
		}
	}
	return colorMap
}
