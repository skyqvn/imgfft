package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/madelynnblue/go-dsp/fft"
)

func cmdRestore(args []string) {
	start := time.Now()
	fs := flag.NewFlagSet("r", flag.ExitOnError)
	in := fs.String("i", "", "input 输入文件（.tiff / .tif / .h5）")
	out := fs.String("o", "out.png", "output 输出 PNG")
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

	switch ext {
	case ".tiff", ".tif":
		chans, tw, th, err := readTIFF(*in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误：读取 TIFF 失败：%v\n", err)
			os.Exit(1)
		}
		if len(chans) != 6 {
			fmt.Fprintf(os.Stderr, "错误：TIFF 必须是 6 通道（当前 %d 通道）\n", len(chans))
			os.Exit(1)
		}
		fftW, fftH = tw, th
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
		spectra, fftW, fftH, err = readHDF5(*in)
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

	outPath := *out
	if !strings.HasSuffix(strings.ToLower(outPath), ".png") {
		outPath += ".png"
	}
	img := mergeChannels(restored[0], restored[1], restored[2], fftW, fftH)
	if err := savePNG(img, outPath); err != nil {
		fmt.Fprintf(os.Stderr, "错误：保存 PNG 失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("输出 -> %s (%dx%d, 耗时 %v)\n", outPath, fftW, fftH, time.Since(start).Round(time.Millisecond))
}
