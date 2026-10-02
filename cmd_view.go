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

func cmdView(args []string) {
	start := time.Now()
	fs := flag.NewFlagSet("v", flag.ExitOnError)
	in := fs.String("i", "", "input 输入图片路径")
	out := fs.String("o", "", "output 输出前缀（默认=输入文件名）")
	mode := fs.String("m", "0", "mode 输出模式：0=幅度+相位（默认）, 1=实部+虚部")
	format := fs.String("f", "png", "format 输出格式：png=8位PNG, tiff=float64 TIFF")
	fs.Parse(args)

	if *in == "" {
		fmt.Fprintln(os.Stderr, "错误：-i 输入路径不能为空")
		fs.Usage()
		os.Exit(1)
	}
	if *mode != "0" && *mode != "1" {
		fmt.Fprintf(os.Stderr, "错误：-m 必须是 0 或 1，当前为 %q\n", *mode)
		os.Exit(1)
	}
	if *format != "png" && *format != "tiff" {
		fmt.Fprintf(os.Stderr, "错误：-f 必须是 png 或 tiff，当前为 %q\n", *format)
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

	prefix := *out
	if prefix == "" {
		prefix = strings.TrimSuffix(*in, filepath.Ext(*in))
	}

	// 模式参数在三个通道间不变，提前判断
	magPhaseMode := *mode == "0"

	var reCh, imCh, magCh, phCh [3][][]float64

	var wg sync.WaitGroup
	for ch := 0; ch < 3; ch++ {
		wg.Add(1)
		go func(ch int) {
			defer wg.Done()
			mat := extractChannel(img, ch)
			spec := fft.FFT2Real(mat)
			fftShiftComplex(spec)

			var re, im, mag, ph [][]float64
			if magPhaseMode {
				_, _, mag, ph = extractComponents(spec, false, false, true, true)
				normalizeMag(mag)
				normalizePh(ph)
			} else {
				re, im, _, _ = extractComponents(spec, true, true, false, false)
				normalizeRealImag(re)
				normalizeRealImag(im)
			}

			reCh[ch] = re
			imCh[ch] = im
			magCh[ch] = mag
			phCh[ch] = ph
		}(ch)
	}
	wg.Wait()

	if magPhaseMode {
		outputPair(prefix, "mag", "phase", magCh, phCh, ow, oh, *format)
	} else {
		outputPair(prefix, "re", "im", reCh, imCh, ow, oh, *format)
	}
	fmt.Printf("完成，耗时 %v。\n", time.Since(start).Round(time.Millisecond))
}

func outputPair(prefix, s1, s2 string, m1, m2 [3][][]float64, w, h int, format string) {
	switch format {
	case "png":
		savePNG(mergeChannels(m1[0], m1[1], m1[2], w, h), prefix+"_"+s1+".png")
		savePNG(mergeChannels(m2[0], m2[1], m2[2], w, h), prefix+"_"+s2+".png")
		fmt.Printf("%s -> %s_%s.png\n", s1, prefix, s1)
		fmt.Printf("%s -> %s_%s.png\n", s2, prefix, s2)
	case "tiff":
		writeViewTIFF(prefix, s1, m1, w, h)
		writeViewTIFF(prefix, s2, m2, w, h)
	}
}

func writeViewTIFF(prefix, name string, chans [3][][]float64, w, h int) {
	flat := [][][]float64{chans[0], chans[1], chans[2]}
	if err := writeTIFF(prefix+"_"+name+".tiff", flat, w, h, 64); err != nil {
		fmt.Fprintf(os.Stderr, "错误：写入 %s 失败：%v\n", name, err)
		os.Exit(1)
	}
	fmt.Printf("%s -> %s_%s.tiff\n", name, prefix, name)
}
