package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madelynnblue/go-dsp/fft"
)

func cmdLossless(args []string) {
	fs := flag.NewFlagSet("l", flag.ExitOnError)
	in := fs.String("i", "", "输入图片路径")
	out := fs.String("o", "", "输出文件（默认=输入文件名.tiff 或 .h5）")
	format := fs.String("f", "hdf5", "输出格式：hdf5=HDF5（默认）, tiff=6通道float64 TIFF")
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
	n := nextPow2(maxInt(ow, oh))
	fmt.Printf("输入 %dx%d，FFT 尺寸 %dx%d\n", ow, oh, n, n)

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
		mat := pad(extractChannel(img, ch), n)
		spectra[ch] = fft.FFT2Real(mat)
	}

	switch *format {
	case "tiff":
		chans := make([][][]float64, 6)
		for ch := 0; ch < 3; ch++ {
			re := make([][]float64, n)
			im := make([][]float64, n)
			for y := 0; y < n; y++ {
				re[y] = make([]float64, n)
				im[y] = make([]float64, n)
				for x := 0; x < n; x++ {
					re[y][x] = real(spectra[ch][y][x])
					im[y][x] = imag(spectra[ch][y][x])
				}
			}
			chans[ch] = re
			chans[ch+3] = im
		}
		if err := writeTIFF(*out, chans, n, n, 64, ow, oh); err != nil {
			fmt.Fprintf(os.Stderr, "错误：写入 TIFF 失败：%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("无损 -> %s（6 通道 float64 TIFF）\n", *out)

	case "hdf5":
		if err := writeHDF5(*out, spectra, ow, oh, n); err != nil {
			fmt.Fprintf(os.Stderr, "错误：写入 HDF5 失败：%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("无损 -> %s（HDF5 float64）\n", *out)
	}
}
