package main

import (
	"fmt"

	"github.com/Komplementariteten/Golang/img-demo/png"
)

func main() {
	png.LoadPngGopherPrintBounds()
	fmt.Println("=====")
	png.LoadPngGopherCountColor()
	fmt.Println("=====")
}
