package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

	// 计算默认输出路径
	prefix := strings.TrimSuffix(*in, filepath.Ext(*in))
	if *out == "" {
		switch *format {
		case "tiff":
			*out = prefix + ".tiff"
		default:
			*out = prefix + ".h5"
		}
	}

	// 三通道并行 FFT
	var spectra [3][][]complex128
	var wg sync.WaitGroup
	for ch := 0; ch < 3; ch++ {
		wg.Add(1)
		go func(ch int) {
			defer wg.Done()
			mat := extractChannel(img, ch)
			spectra[ch] = fft.FFT2Real(mat)
		}(ch)
	}
	wg.Wait()

	switch *format {
	case "tiff":
		saveAsTIFF(*out, spectra, ow, oh, start)
	case "hdf5":
		saveAsHDF5(*out, spectra, ow, oh, start)
	}
}

// =============================================================================
// save 子命令：按格式输出
// =============================================================================

func saveAsTIFF(outPath string, spectra [3][][]complex128, w, h int, start time.Time) {
	chans := make([][][]float64, 6)
	var wg sync.WaitGroup
	for ch := 0; ch < 3; ch++ {
		wg.Add(1)
		go func(ch int) {
			defer wg.Done()
			re, im := splitComplexToFloat64(spectra[ch], w, h)
			chans[ch] = re
			chans[ch+3] = im
		}(ch)
	}
	wg.Wait()
	if err := writeTIFF(outPath, chans, w, h, 64); err != nil {
		fmt.Fprintf(os.Stderr, "错误：写入 TIFF 失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("无损 -> %s（6 通道 float64 TIFF，耗时 %v）\n", outPath, time.Since(start).Round(time.Millisecond))
}

func saveAsHDF5(outPath string, spectra [3][][]complex128, w, h int, start time.Time) {
	if err := writeHDF5(outPath, spectra, w, h); err != nil {
		fmt.Fprintf(os.Stderr, "错误：写入 HDF5 失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("无损 -> %s（HDF5 float64，耗时 %v）\n", outPath, time.Since(start).Round(time.Millisecond))
}

// splitComplexToFloat64 将复数频谱拆分为实部/虚部两个 [][]float64 矩阵。
func splitComplexToFloat64(spec [][]complex128, w, h int) (re, im [][]float64) {
	reData := make([]float64, h*w)
	imData := make([]float64, h*w)
	re = make([][]float64, h)
	im = make([][]float64, h)
	for y := 0; y < h; y++ {
		re[y] = reData[y*w : (y+1)*w]
		im[y] = imData[y*w : (y+1)*w]
		for x := 0; x < w; x++ {
			re[y][x] = real(spec[y][x])
			im[y][x] = imag(spec[y][x])
		}
	}
	return
}
