package img_util

import (
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"ventose.cc/img"
)

func PaintToScreen(rgba *img.RgbImage) error {
	rgbaImage := rgba.Paint()
	app := app.New()
	w := app.NewWindow("Image Pain")
	w.SetContent(
		canvas.NewImageFromImage(rgbaImage),
		)
	w.ShowAndRun()
}
