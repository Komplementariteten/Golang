package img

import (
	"image/color"
	"math"
	"testing"
)

func TestRgbImage_SetNoice(t *testing.T) {
	img, _ := NewImage(100, 100)
	red := &color.RGBA{A:255, R:255, B:0, G:0}
	err := img.SetNoice(red, 0.2)
	if err != nil {
		t.Error(err)
	}
	colValue := ColorToUint(red)
	colPixCount := int(math.Floor(float64(len(img.Raw)) * 0.2))
	foundColPix := 0
	for i := 0; i < len(img.Raw); i++ {
		if img.Raw[i] == colValue {
			foundColPix++
		}
	}
	if colPixCount < foundColPix && foundColPix > 0 {
		t.Errorf("Color Pixel don't match expected found: %d, expected: %d", foundColPix, colPixCount)
	}
}

func TestRgbImage_Fill(t *testing.T) {
	img, _ := NewImage(10, 10)
	red := &color.RGBA{A:255, R:255, B:0, G:0}
	img.Fill(red)

	rgbaImg := img.Paint()
	testColor := rgbaImg.At(5, 5).(color.RGBA)
	if testColor.G != 0 {
		t.Error("Green found in Red filled test Image")
	}

	if testColor.B != 0 {
		t.Error("Blue found in Red filled test Image")
	}

	if testColor.R != 0xFF {
		t.Error("not enough Red found in Red filled test Image")
	}
}

func TestRgbImage_Paint(t *testing.T) {
	img, _ := NewImage(100, 100)
	img.Fill(&color.RGBA{G:255, R:255, A:255, B:255})
	img.Paint()
}

func TestColorToUint(t *testing.T) {
	orgColor := &color.RGBA{0x10, 0x20, 0x30, 0x40}
	uintValue := ColorToUint(orgColor)
	if uintValue != 0x10203040 {
		t.Error("Uint is not 0x10203040 (270544960)")
	}
	testColor := UintToColor(uintValue)

	if testColor.A != orgColor.A {
		t.Error("Alpha value don't match")
	}

	if testColor.B != orgColor.B {
		t.Error("Blue value don't match")
	}

	if testColor.R != orgColor.R {
		t.Error("Red value don't match")
	}

	if testColor.G != orgColor.G {
		t.Error("Green value don't match")
	}
}

func TestUintToPixel(t *testing.T) {
	nearBlack := uint32(0x01010101)
	p := UintToColor(nearBlack)
	if p.A != 1 {
		t.Error("Alpha 0x01 should be 1 dec")
	}
	if p.B != 1 {
		t.Error("Blue 0x01 should be 1 dec")
	}
	if p.G != 1 {
		t.Error("Green 0x01 should be 1 dec")
	}
	if p.R != 1 {
		t.Error("Red 0x01 should be 1 dec")
	}

	white := uint32(0xFFFFFFFF)
	p = UintToColor(white)
	if p.A != 255 {
		t.Error("Alpha 0xFF should be 255 dec")
	}
	if p.B != 255 {
		t.Error("Blue 0xFF should be 255 dec")
	}
	if p.G != 255 {
		t.Error("Green 0xFF should be 255 dec")
	}
	if p.R != 255 {
		t.Error("Red 0xFF should be 255 dec")
	}

	color := uint32(0xD0C0B0A0)
	p = UintToColor(color)
	if p.A != 160 {
		t.Error("Alpha 0xFF should be 255 dec")
	}
	if p.B != 176 {
		t.Error("Blue 0xFF should be 255 dec")
	}
	if p.G != 192 {
		t.Error("Green 0xFF should be 255 dec")
	}
	if p.R != 208 {
		t.Error("Red 0xFF should be 255 dec")
	}

}
