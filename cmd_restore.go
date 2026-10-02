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

	// 1. 读取频谱
	spectra, fftW, fftH := loadSpectra(*in, ext)

	// 2. 三通道并行 IFFT
	restored := restoreChannels(spectra, fftW, fftH)

	// 3. 输出 PNG
	outPath := ensurePNGExt(*out)
	img := mergeChannels(restored[0], restored[1], restored[2], fftW, fftH)
	if err := savePNG(img, outPath); err != nil {
		fmt.Fprintf(os.Stderr, "错误：保存 PNG 失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("输出 -> %s (%dx%d, 耗时 %v)\n", outPath, fftW, fftH, time.Since(start).Round(time.Millisecond))
}

// =============================================================================
// restore 子命令：频谱加载
// =============================================================================

func loadSpectra(inPath, ext string) ([3][][]complex128, int, int) {
	switch ext {
	case ".tiff", ".tif":
		chans, w, h, err := readTIFF(inPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误：读取 TIFF 失败：%v\n", err)
			os.Exit(1)
		}
		if len(chans) != 6 {
			fmt.Fprintf(os.Stderr, "错误：TIFF 必须是 6 通道（当前 %d 通道）\n", len(chans))
			os.Exit(1)
		}
		return combineReIm(chans, w, h), w, h
	default: // .h5
		s, w, h, err := readHDF5(inPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误：读取 HDF5 失败：%v\n", err)
			os.Exit(1)
		}
		return s, w, h
	}
}

// combineReIm 从 6 通道 [re,re,re, im,im,im] 合并为 3 通道 [][]complex128。
func combineReIm(chans [][][]float64, w, h int) [3][][]complex128 {
	var spectra [3][][]complex128
	var wg sync.WaitGroup
	for ch := 0; ch < 3; ch++ {
		wg.Add(1)
		go func(ch int) {
			defer wg.Done()
			specData := make([]complex128, h*w)
			spectra[ch] = make([][]complex128, h)
			for y := 0; y < h; y++ {
				spectra[ch][y] = specData[y*w : (y+1)*w]
				for x := 0; x < w; x++ {
					spectra[ch][y][x] = complex(chans[ch][y][x], chans[ch+3][y][x])
				}
			}
		}(ch)
	}
	wg.Wait()
	return spectra
}

// =============================================================================
// restore 子命令：逆变换与输出
// =============================================================================

func restoreChannels(spectra [3][][]complex128, w, h int) [3][][]float64 {
	var restored [3][][]float64
	var wg sync.WaitGroup
	for ch := 0; ch < 3; ch++ {
		wg.Add(1)
		go func(ch int) {
			defer wg.Done()
			res := fft.IFFT2(spectra[ch])
			matData := make([]float64, h*w)
			mat := make([][]float64, h)
			for y := 0; y < h; y++ {
				mat[y] = matData[y*w : (y+1)*w]
				for x := 0; x < w; x++ {
					mat[y][x] = real(res[y][x])
				}
			}
			restored[ch] = mat
		}(ch)
	}
	wg.Wait()
	return restored
}

func ensurePNGExt(path string) string {
	if strings.HasSuffix(strings.ToLower(path), ".png") {
		return path
	}
	return path + ".png"
}
