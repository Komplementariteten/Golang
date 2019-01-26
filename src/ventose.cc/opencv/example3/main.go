package main

import (
	"fmt"
	"github.com/lazywei/go-opencv/opencv"
)

func main() {
	window := opencv.NewWindow("Sample opencv Window")
	defer window.Destroy()

	cap := opencv.NewCameraCapture(0)
	if cap == nil {
		panic("Can't open Camera")
	}
	defer cap.Release()

	fmt.Println("Press ESC to quit")

	for {
		if cap.GrabFrame() {
			img := cap.RetrieveFrame(1)
			if img != nil {
				grayImg := opencv.CreateImage(img.Width(), img.Height(), opencv.IPL_DEPTH_8U, 1)
				binImg := opencv.CreateImage(img.Width(), img.Height(), opencv.IPL_DEPTH_8U, 1)
				defer grayImg.Release()
				defer binImg.Release()
			}
		}
	}

}
