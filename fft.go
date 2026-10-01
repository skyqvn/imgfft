package main

import (
	"math"
	"math/cmplx"
)

func nextPow2(n int) int {
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func pad(mat [][]float64, size int) [][]float64 {
	h := len(mat)
	out := make([][]float64, size)
	for y := 0; y < size; y++ {
		out[y] = make([]float64, size)
		if y < h {
			copy(out[y], mat[y])
		}
	}
	return out
}

func fftShiftComplex(spec [][]complex128) [][]complex128 {
	h, w := len(spec), len(spec[0])
	out := make([][]complex128, h)
	for y := 0; y < h; y++ {
		out[y] = make([]complex128, w)
		for x := 0; x < w; x++ {
			out[y][x] = spec[(y+h/2)%h][(x+w/2)%w]
		}
	}
	return out
}

func extractComponents(spec [][]complex128) (re, im, mag, ph [][]float64) {
	n := len(spec)
	re = make([][]float64, n)
	im = make([][]float64, n)
	mag = make([][]float64, n)
	ph = make([][]float64, n)
	for y := 0; y < n; y++ {
		re[y] = make([]float64, n)
		im[y] = make([]float64, n)
		mag[y] = make([]float64, n)
		ph[y] = make([]float64, n)
		for x := 0; x < n; x++ {
			c := spec[y][x]
			re[y][x] = real(c)
			im[y][x] = imag(c)
			mag[y][x] = cmplx.Abs(c)
			ph[y][x] = cmplx.Phase(c)
		}
	}
	return
}

func logCompress(mat [][]float64, symmetric bool) {
	var mx float64
	for y := range mat {
		for x := range mat[y] {
			if a := math.Abs(mat[y][x]); a > mx {
				mx = a
			}
		}
	}
	if mx == 0 {
		mx = 1
	}
	invLog := 1.0 / math.Log1p(mx)
	for y := range mat {
		for x := range mat[y] {
			v := mat[y][x]
			lv := math.Log1p(math.Abs(v)) * invLog
			if symmetric {
				if v < 0 {
					lv = -lv
				}
				mat[y][x] = (lv + 1) / 2 * 255
			} else {
				mat[y][x] = lv * 255
			}
		}
	}
}

func normalizeRealImag(mat [][]float64) { logCompress(mat, true) }

func normalizeMag(mag [][]float64) { logCompress(mag, false) }

func normalizePh(ph [][]float64) {
	for y := range ph {
		for x := range ph[y] {
			ph[y][x] = (ph[y][x] + math.Pi) / (2 * math.Pi) * 255
		}
	}
}
