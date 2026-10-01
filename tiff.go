package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

const (
	tiffTagImageWidth      = 256
	tiffTagImageLength     = 257
	tiffTagBitsPerSample   = 258
	tiffTagCompression     = 259
	tiffTagPhotometric     = 262
	tiffTagStripOffsets    = 273
	tiffTagSamplesPerPixel = 277
	tiffTagRowsPerStrip    = 278
	tiffTagStripByteCounts = 279
	tiffTagPlanarConfig    = 284
	tiffTagExtraSamples    = 338
	tiffTagSampleFormat    = 339
)

type tiffEntry struct {
	tag, typ uint16
	count    uint32
	value    uint32
}

func writeTIFF(path string, chans [][][]float64, w, h int, bits int) error {
	numCh := len(chans)
	if numCh != 3 && numCh != 6 {
		return fmt.Errorf("TIFF 只支持 3 或 6 通道，当前 %d 通道", numCh)
	}
	if bits != 32 && bits != 64 {
		return fmt.Errorf("TIFF 位深必须是 32 或 64，当前 %d", bits)
	}

	bytesPerSample := bits / 8
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write([]byte{'I', 'I'})
	binary.Write(f, binary.LittleEndian, uint16(42))
	binary.Write(f, binary.LittleEndian, uint32(8))

	numEntries := uint16(11)
	if numCh == 6 {
		numEntries = 12
	}
	ifdSize := uint32(2) + uint32(numEntries)*12 + 4
	bitsOffset := uint32(8) + ifdSize
	sampleFmtOffset := bitsOffset + uint32(numCh)*2
	var extraOffset uint32
	var pixelOffset uint32
	if numCh == 6 {
		extraOffset = sampleFmtOffset + uint32(numCh)*2
		pixelOffset = extraOffset + 6
	} else {
		pixelOffset = sampleFmtOffset + uint32(numCh)*2
	}

	binary.Write(f, binary.LittleEndian, numEntries)

	entries := []tiffEntry{
		{tiffTagImageWidth, 4, 1, uint32(w)},
		{tiffTagImageLength, 4, 1, uint32(h)},
		{tiffTagBitsPerSample, 3, uint32(numCh), bitsOffset},
		{tiffTagCompression, 3, 1, 1},
		{tiffTagPhotometric, 3, 1, 2},
		{tiffTagStripOffsets, 4, 1, pixelOffset},
		{tiffTagSamplesPerPixel, 3, 1, uint32(numCh)},
		{tiffTagRowsPerStrip, 4, 1, uint32(h)},
		{tiffTagStripByteCounts, 4, 1, uint32(w * h * numCh * bytesPerSample)},
		{tiffTagPlanarConfig, 3, 1, 1},
		{tiffTagSampleFormat, 3, uint32(numCh), sampleFmtOffset},
	}
	if numCh == 6 {
		entries = append(entries, tiffEntry{tiffTagExtraSamples, 3, 3, extraOffset})
	}
	for _, e := range entries {
		binary.Write(f, binary.LittleEndian, e.tag)
		binary.Write(f, binary.LittleEndian, e.typ)
		binary.Write(f, binary.LittleEndian, e.count)
		binary.Write(f, binary.LittleEndian, e.value)
	}
	binary.Write(f, binary.LittleEndian, uint32(0))

	for i := 0; i < numCh; i++ {
		binary.Write(f, binary.LittleEndian, uint16(bits))
	}
	for i := 0; i < numCh; i++ {
		binary.Write(f, binary.LittleEndian, uint16(3))
	}
	if numCh == 6 {
		for i := 0; i < 3; i++ {
			binary.Write(f, binary.LittleEndian, uint16(0))
		}
	}

	buf := make([]byte, w*h*numCh*bytesPerSample)
	bo := binary.LittleEndian
	idx := 0
	if bits == 32 {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				for ch := 0; ch < numCh; ch++ {
					bo.PutUint32(buf[idx:], math.Float32bits(float32(chans[ch][y][x])))
					idx += 4
				}
			}
		}
	} else {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				for ch := 0; ch < numCh; ch++ {
					bo.PutUint64(buf[idx:], math.Float64bits(chans[ch][y][x]))
					idx += 8
				}
			}
		}
	}
	f.Write(buf)
	return nil
}

func readTIFF(path string) ([][][]float64, int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, err
	}
	defer f.Close()

	var bo [2]byte
	io.ReadFull(f, bo[:])
	var order binary.ByteOrder = binary.LittleEndian
	if bo[0] == 'M' {
		order = binary.BigEndian
	}

	var magic uint16
	binary.Read(f, order, &magic)
	if magic != 42 {
		return nil, 0, 0, fmt.Errorf("不是有效的 TIFF 文件")
	}

	var ifdOffset uint32
	binary.Read(f, order, &ifdOffset)
	f.Seek(int64(ifdOffset), io.SeekStart)

	var numEntries uint16
	binary.Read(f, order, &numEntries)

	var w, h, numCh, stripOffset uint32
	var bitsPerSample uint32
	var bitsOffset uint32

	for i := 0; i < int(numEntries); i++ {
		var e tiffEntry
		binary.Read(f, order, &e.tag)
		binary.Read(f, order, &e.typ)
		binary.Read(f, order, &e.count)
		binary.Read(f, order, &e.value)
		switch e.tag {
		case tiffTagImageWidth:
			w = e.value
		case tiffTagImageLength:
			h = e.value
		case tiffTagSamplesPerPixel:
			numCh = e.value
		case tiffTagStripOffsets:
			stripOffset = e.value
		case tiffTagBitsPerSample:
			if e.count == 1 {
				bitsPerSample = e.value
			} else {
				bitsOffset = e.value
			}
		}
	}

	if bitsPerSample == 0 && bitsOffset != 0 {
		cur, _ := f.Seek(0, io.SeekCurrent)
		f.Seek(int64(bitsOffset), io.SeekStart)
		var bps uint16
		binary.Read(f, order, &bps)
		bitsPerSample = uint32(bps)
		f.Seek(cur, io.SeekStart)
	}

	if numCh != 3 && numCh != 6 {
		return nil, 0, 0, fmt.Errorf("TIFF 通道数必须是 3 或 6，当前 %d", numCh)
	}
	if w == 0 || h == 0 {
		return nil, 0, 0, fmt.Errorf("TIFF 尺寸无效：%dx%d", w, h)
	}
	if bitsPerSample != 32 && bitsPerSample != 64 {
		return nil, 0, 0, fmt.Errorf("TIFF 位深必须是 32 或 64，当前 %d", bitsPerSample)
	}

	f.Seek(int64(stripOffset), io.SeekStart)
	chans := make([][][]float64, numCh)
	for ch := 0; ch < int(numCh); ch++ {
		chans[ch] = make([][]float64, h)
		for y := 0; y < int(h); y++ {
			chans[ch][y] = make([]float64, w)
		}
	}

	pixelBytes := int(w) * int(h) * int(numCh) * int(bitsPerSample) / 8
	buf := make([]byte, pixelBytes)
	if _, err := io.ReadFull(f, buf); err != nil {
		return nil, 0, 0, fmt.Errorf("读取像素数据失败: %w", err)
	}

	idx := 0
	switch bitsPerSample {
	case 32:
		for y := 0; y < int(h); y++ {
			for x := 0; x < int(w); x++ {
				for ch := 0; ch < int(numCh); ch++ {
					bits := order.Uint32(buf[idx:])
					chans[ch][y][x] = float64(math.Float32frombits(bits))
					idx += 4
				}
			}
		}
	case 64:
		for y := 0; y < int(h); y++ {
			for x := 0; x < int(w); x++ {
				for ch := 0; ch < int(numCh); ch++ {
					bits := order.Uint64(buf[idx:])
					chans[ch][y][x] = math.Float64frombits(bits)
					idx += 8
				}
			}
		}
	}
	return chans, int(w), int(h), nil
}
