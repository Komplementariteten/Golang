package png

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"os"
)

func loadPngGopher() (image.Image, image.Rectangle) {
	reader, err := os.Open("./example/gopher.png")
	if err != nil {
		reader, err = os.Open("../../example/gopher.png")
		if err != nil {
			log.Fatal(err)
		}
	}

	m, _, err := image.Decode(reader)

	if err != nil {
		log.Fatal(err)
	}

	return m, m.Bounds()
}

func LoadPngGopherCountColor() {
	img, bounds := loadPngGopher()
	colorMap := make(map[color.Color]int)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			color := img.At(x, y)
			colorMap[color]++
		}
	}

	for index := range colorMap {
		if colorMap[index] > 1 {
			fmt.Printf("%s => %d\n", index, colorMap[index])
		}
	}

}

func LoadPngGopherPrintBounds() {
	fmt.Println(os.Getwd())
	m, bounds := loadPngGopher()
	cm := m.ColorModel()
	fmt.Printf("min: %v, %v max: %v, %v \n", bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
	fmt.Printf("Color model: %s\n", cm)
}
