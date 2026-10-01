package main

import (
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

func clamp255(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func savePNG(img image.Image, path string) error {
	if !strings.HasSuffix(strings.ToLower(path), ".png") {
		path = path + ".png"
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func extractChannel(img image.Image, ch int) [][]float64 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	mat := make([][]float64, h)
	for y := 0; y < h; y++ {
		mat[y] = make([]float64, w)
		for x := 0; x < w; x++ {
			r, g, bb, _ := img.At(x+b.Min.X, y+b.Min.Y).RGBA()
			var v uint32
			switch ch {
			case 0:
				v = r
			case 1:
				v = g
			case 2:
				v = bb
			}
			mat[y][x] = float64(v >> 8)
		}
	}
	return mat
}

func mergeChannels(r, g, b [][]float64, w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{
				clamp255(r[y][x]), clamp255(g[y][x]), clamp255(b[y][x]), 255,
			})
		}
	}
	return img
}
