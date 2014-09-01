package main

import (
	"fmt"
	"github.com/rubyist/gohat/pkg/heapfile"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

var width = 1024
var blue = color.RGBA{0, 0, 255, 255}
var red = color.RGBA{255, 0, 0, 255}

func main() {

	if len(os.Args) != 1 {
		fmt.Println("gohatr <heapfile>")
		os.Exit(1)
	}

	heapFile, err := heapfile.New(os.Args[1])
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	f, err := os.Create("image.png")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	heapStart := heapFile.DumpParams().StartAddress
	heapEnd := heapFile.DumpParams().EndAddress
	rows := int(heapEnd-heapStart) / width // might need to check %

	img := image.NewRGBA(image.Rect(0, 0, width, rows))
	draw.Draw(img, img.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)

	for _, object := range heapFile.Objects() {
		offset := int(object.Address - heapStart)
		drawObject(img, offset, object.Size)
	}

	png.Encode(f, img)
	f.Close()
}

func drawObject(img draw.Image, offset, size int) {
	ay := offset / width
	ax := offset % width
	dx := minInt(ax+size, width)
	dy := ay + 1
	left := size - (dx - ax)

	draw.Draw(img, image.Rect(ax, ay, dx, dy), &image.Uniform{red}, image.ZP, draw.Src)

	if left > 0 {
		drawn := dx - ax
		drawObject(img, offset+drawn, left)
	}
}
