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

func cmdSave(args []string) {
	start := time.Now()
	fs := flag.NewFlagSet("s", flag.ExitOnError)
	in := fs.String("i", "", "input 输入图片路径")
	out := fs.String("o", "", "output 输出文件（默认=输入文件名.h5）")
	format := fs.String("f", "hdf5", "format 输出格式：hdf5=HDF5（默认）, tiff=6通道float64 TIFF")
	fs.Parse(args)

	if *in == "" {
		fmt.Fprintln(os.Stderr, "错误：-i 输入路径不能为空")
		fs.Usage()
		os.Exit(1)
	}
	if *format != "tiff" && *format != "hdf5" {
		fmt.Fprintf(os.Stderr, "错误：-f 必须是 tiff 或 hdf5，当前为 %q\n", *format)
		os.Exit(1)
	}

	img, err := loadImage(*in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误：加载图片失败：%v\n", err)
		os.Exit(1)
	}

	b := img.Bounds()
	ow, oh := b.Dx(), b.Dy()
	fmt.Printf("输入 %dx%d\n", ow, oh)

	prefix := strings.TrimSuffix(*in, filepath.Ext(*in))
	if *out == "" {
		if *format == "tiff" {
			*out = prefix + ".tiff"
		} else {
			*out = prefix + ".h5"
		}
	}

	var spectra [3][][]complex128
	for ch := 0; ch < 3; ch++ {
		mat := extractChannel(img, ch)
		spectra[ch] = fft.FFT2Real(mat)
	}

	switch *format {
	case "tiff":
		chans := make([][][]float64, 6)
		for ch := 0; ch < 3; ch++ {
			re := make([][]float64, oh)
			im := make([][]float64, oh)
			for y := 0; y < oh; y++ {
				re[y] = make([]float64, ow)
				im[y] = make([]float64, ow)
				for x := 0; x < ow; x++ {
					re[y][x] = real(spectra[ch][y][x])
					im[y][x] = imag(spectra[ch][y][x])
				}
			}
			chans[ch] = re
			chans[ch+3] = im
		}
		if err := writeTIFF(*out, chans, ow, oh, 64); err != nil {
			fmt.Fprintf(os.Stderr, "错误：写入 TIFF 失败：%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("无损 -> %s（6 通道 float64 TIFF，耗时 %v）\n", *out, time.Since(start).Round(time.Millisecond))

	case "hdf5":
		if err := writeHDF5(*out, spectra, ow, oh); err != nil {
			fmt.Fprintf(os.Stderr, "错误：写入 HDF5 失败：%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("无损 -> %s（HDF5 float64，耗时 %v）\n", *out, time.Since(start).Round(time.Millisecond))
	}
}
