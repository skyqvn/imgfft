package main

import (
	"fmt"
	"math"

	"github.com/scigolib/hdf5"
)

func writeHDF5(path string, spectra [3][][]complex128, ow, oh, n int) error {
	fw, err := hdf5.CreateForWrite(path, hdf5.CreateTruncate)
	if err != nil {
		return fmt.Errorf("创建 HDF5 文件失败: %w", err)
	}

	ds, err := fw.CreateDataset("/fft", hdf5.Float64, []uint64{3, uint64(n), uint64(n), 2})
	if err != nil {
		fw.Close()
		return fmt.Errorf("创建数据集失败: %w", err)
	}

	total := 3 * n * n * 2
	flat := make([]float64, total)
	idx := 0
	for ch := 0; ch < 3; ch++ {
		for y := 0; y < n; y++ {
			for x := 0; x < n; x++ {
				flat[idx] = real(spectra[ch][y][x])
				flat[idx+1] = imag(spectra[ch][y][x])
				idx += 2
			}
		}
	}

	if err := ds.Write(flat); err != nil {
		fw.Close()
		return fmt.Errorf("写入数据失败: %w", err)
	}

	if err := ds.WriteAttribute("orig_size", []int32{int32(ow), int32(oh)}); err != nil {
		fw.Close()
		return fmt.Errorf("写入属性失败: %w", err)
	}

	if err := fw.Close(); err != nil {
		return fmt.Errorf("关闭文件失败: %w", err)
	}
	return nil
}

func readHDF5(path string) ([3][][]complex128, int, int, int, int, error) {
	var zero [3][][]complex128

	f, err := hdf5.Open(path)
	if err != nil {
		return zero, 0, 0, 0, 0, fmt.Errorf("打开 HDF5 文件失败: %w", err)
	}
	defer f.Close()

	var ds *hdf5.Dataset
	f.Walk(func(walkPath string, obj hdf5.Object) {
		if walkPath == "/fft" || walkPath == "fft" {
			if d, ok := obj.(*hdf5.Dataset); ok {
				ds = d
			}
		}
	})

	if ds == nil {
		return zero, 0, 0, 0, 0, fmt.Errorf("未找到 /fft 数据集")
	}

	data, err := ds.Read()
	if err != nil {
		return zero, 0, 0, 0, 0, fmt.Errorf("读取数据集失败: %w", err)
	}

	rawOrig, err := ds.ReadAttribute("orig_size")
	if err != nil {
		return zero, 0, 0, 0, 0, fmt.Errorf("读取 orig_size 属性失败: %w", err)
	}

	origSlice, ok := rawOrig.([]int32)
	if !ok || len(origSlice) < 2 {
		return zero, 0, 0, 0, 0, fmt.Errorf("orig_size 属性格式错误: %T", rawOrig)
	}
	ow, oh := int(origSlice[0]), int(origSlice[1])

	nSq := len(data)
	n := int(math.Sqrt(float64(nSq / (3 * 2))))
	if 3*n*n*2 != nSq {
		return zero, 0, 0, 0, 0, fmt.Errorf("数据大小异常: %d 个 float64, 非 3*n*n*2 格式", nSq)
	}

	var spectra [3][][]complex128
	idx := 0
	for ch := 0; ch < 3; ch++ {
		spectra[ch] = make([][]complex128, n)
		for y := 0; y < n; y++ {
			spectra[ch][y] = make([]complex128, n)
			for x := 0; x < n; x++ {
				spectra[ch][y][x] = complex(data[idx], data[idx+1])
				idx += 2
			}
		}
	}

	return spectra, n, n, ow, oh, nil
}
