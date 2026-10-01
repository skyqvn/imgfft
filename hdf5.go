package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/scigolib/hdf5"
)

func writeHDF5(path string, spectra [3][][]complex128, w, h int) error {
	fw, err := hdf5.CreateForWrite(path, hdf5.CreateTruncate)
	if err != nil {
		return fmt.Errorf("创建 HDF5 文件失败: %w", err)
	}

	ds, err := fw.CreateDataset("/fft", hdf5.Float64, []uint64{3, uint64(h), uint64(w), 2})
	if err != nil {
		fw.Close()
		return fmt.Errorf("创建数据集失败: %w", err)
	}

	total := 3 * h * w * 2
	flat := make([]float64, total)
	idx := 0
	for ch := 0; ch < 3; ch++ {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
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

	if err := fw.Close(); err != nil {
		return fmt.Errorf("关闭文件失败: %w", err)
	}
	return nil
}

func readHDF5(path string) ([3][][]complex128, int, int, error) {
	var zero [3][][]complex128

	f, err := hdf5.Open(path)
	if err != nil {
		return zero, 0, 0, fmt.Errorf("打开 HDF5 文件失败: %w", err)
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
		return zero, 0, 0, fmt.Errorf("未找到 /fft 数据集")
	}

	info, err := ds.Info()
	if err != nil {
		return zero, 0, 0, fmt.Errorf("读取数据集信息失败: %w", err)
	}

	w, h, err := parseDims(info)
	if err != nil {
		return zero, 0, 0, fmt.Errorf("解析维度失败: %w", err)
	}

	data, err := ds.Read()
	if err != nil {
		return zero, 0, 0, fmt.Errorf("读取数据集失败: %w", err)
	}

	if 3*w*h*2 != len(data) {
		return zero, 0, 0, fmt.Errorf("数据大小异常: %d 个 float64, 期望 %d", len(data), 3*w*h*2)
	}

	var spectra [3][][]complex128
	idx := 0
	for ch := 0; ch < 3; ch++ {
		spectra[ch] = make([][]complex128, h)
		for y := 0; y < h; y++ {
			spectra[ch][y] = make([]complex128, w)
			for x := 0; x < w; x++ {
				spectra[ch][y][x] = complex(data[idx], data[idx+1])
				idx += 2
			}
		}
	}

	return spectra, w, h, nil
}

// parseDims 从 Info() 返回的字符串中提取 w 和 h。
// Info() 格式示例: "Dataset: Float64, 4D array [3 1080 1920 2], Contiguous"
// dims 顺序为 [channels, height, width, 2]，取 parts[2]=width, parts[1]=height。
func parseDims(info string) (w, h int, err error) {
	start := strings.Index(info, "[")
	if start == -1 {
		return 0, 0, fmt.Errorf("未找到维度信息 [%s", info)
	}
	end := strings.Index(info[start:], "]")
	if end == -1 {
		return 0, 0, fmt.Errorf("未找到维度结尾 ]: %s", info)
	}
	rest := info[start+1 : start+end]

	parts := strings.Fields(rest)
	if len(parts) != 4 {
		return 0, 0, fmt.Errorf("期望 4 维，实际 %d 维: %s", len(parts), info)
	}

	w64, e1 := strconv.ParseUint(parts[2], 10, 64)
	h64, e2 := strconv.ParseUint(parts[1], 10, 64)
	if e1 != nil || e2 != nil {
		return 0, 0, fmt.Errorf("解析维度数字失败: %s", info)
	}
	return int(w64), int(h64), nil
}
