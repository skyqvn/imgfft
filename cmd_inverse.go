package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madelynnblue/go-dsp/fft"
)

func cmdInverse(args []string) {
	fs := flag.NewFlagSet("i", flag.ExitOnError)
	in := fs.String("i", "", "输入文件（.tiff / .tif / .h5）")
	out := fs.String("o", "out.png", "输出 PNG")
	fs.Parse(args)

	if *in == "" {
		fmt.Fprintln(os.Stderr, "错误：-i 输入路径不能为空")
		fs.Usage()
		os.Exit(1)
	}

	ext := strings.ToLower(filepath.Ext(*in))
	if ext != ".tiff" && ext != ".tif" && ext != ".h5" {
		fmt.Fprintf(os.Stderr, "错误：输入文件必须是 .tiff / .tif / .h5，当前为 %q\n", ext)
		os.Exit(1)
	}

	var spectra [3][][]complex128
	var fftW, fftH int
	var origW, origH int

	switch ext {
	case ".tiff", ".tif":
		chans, tw, th, tow, toh, err := readTIFF(*in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误：读取 TIFF 失败：%v\n", err)
			os.Exit(1)
		}
		if len(chans) != 6 {
			fmt.Fprintf(os.Stderr, "错误：TIFF 必须是 6 通道（当前 %d 通道）\n", len(chans))
			os.Exit(1)
		}
		fftW, fftH = tw, th
		origW, origH = tow, toh
		for ch := 0; ch < 3; ch++ {
			spectra[ch] = make([][]complex128, fftH)
			for y := 0; y < fftH; y++ {
				spectra[ch][y] = make([]complex128, fftW)
				for x := 0; x < fftW; x++ {
					spectra[ch][y][x] = complex(chans[ch][y][x], chans[ch+3][y][x])
				}
			}
		}
	case ".h5":
		var err error
		spectra, fftW, fftH, origW, origH, err = readHDF5(*in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误：读取 HDF5 失败：%v\n", err)
			os.Exit(1)
		}
	}

	var restored [3][][]float64
	for ch := 0; ch < 3; ch++ {
		res := fft.IFFT2(spectra[ch])
		mat := make([][]float64, fftH)
		for y := 0; y < fftH; y++ {
			mat[y] = make([]float64, fftW)
			for x := 0; x < fftW; x++ {
				mat[y][x] = real(res[y][x])
			}
		}
		restored[ch] = mat
	}

	cropped := make([][][]float64, 3)
	for ch := 0; ch < 3; ch++ {
		cropped[ch] = make([][]float64, origH)
		for y := 0; y < origH; y++ {
			cropped[ch][y] = make([]float64, origW)
			for x := 0; x < origW; x++ {
				cropped[ch][y][x] = restored[ch][y][x]
			}
		}
	}

	img := mergeChannels(cropped[0], cropped[1], cropped[2], origW, origH)
	if err := savePNG(img, *out); err != nil {
		fmt.Fprintf(os.Stderr, "错误：保存 PNG 失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("输出 -> %s (%dx%d)\n", *out, origW, origH)
}
