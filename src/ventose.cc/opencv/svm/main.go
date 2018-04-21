package main

import (
	"github.com/lazywei/go-opencv/opencv"
)

const (
	WIDTH = 512
	HEIGHT = 512
	NTRAINING_SAMPLE = 100
	FRAC_LINEAR_SEP = 0.9
)

func main() {

	// visual representation
	I := opencv.CreateMat(WIDTH, HEIGHT, opencv.CV_8U)

	// -- Train Data ---
	opencv.M
	trainData := opencv.CreateMat(2 * NTRAINING_SAMPLE, 2, opencv.CV_32F)
	labels := opencv.CreateMat(2 * NTRAINING_SAMPLE, 1, opencv.CV_32F)

	rng := opencv.RNG(100)

	nLinearSamples := int(NTRAINING_SAMPLE * NTRAINING_SAMPLE)

	trainClass := trainData.GetRows()
}

