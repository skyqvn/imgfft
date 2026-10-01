# imgfft

图像二维傅里叶变换工具——对 RGB 图片分三通道做 2D FFT，支持可视化频谱输出与精确逆变换还原。

## 安装

```bash
go install github.com/skyqvn/imgfft@latest
```

或克隆后编译：

```bash
git clone https://github.com/skyqvn/imgfft.git
cd imgfft
go build
```

## 快速开始

```bash
# 有损正向变换：生成幅度图 / 相位图（默认）
imgfft f -i photo.png

# 无损正向变换：保存完整复数频谱（默认 HDF5）
imgfft l -i photo.png

# 逆变换：从无损文件精确还原原图
imgfft i -i photo.h5 -o restored.png
```

## 命令

### `f` — 有损正向变换

输出可供人眼直接查看的频谱图，**不可逆变换还原**。

| 选项 | 说明                                        | 默认值     |
| ---- | ------------------------------------------- | ---------- |
| `-i` | 输入图片路径（必需）                        | —          |
| `-o` | 输出前缀                                    | 输入文件名 |
| `-m` | 输出模式：`0`=幅度+相位，`1`=实部+虚部      | `0`        |
| `-t` | 输出格式：`png`=8位PNG，`tiff`=float64 TIFF | `png`      |

输出文件命名规则：
- 模式 0：`<前缀>_mag.png` `"<前缀>_phase.png"`
- 模式 1：`<前缀>_re.png` `"<前缀>_im.png"`

FFT 尺寸自动取不小于 max(宽, 高) 的下一个 2 的幂。

### `l` — 无损正向变换

保存完整复数频谱（float64），**配合 `i` 命令可精确还原原图**。原图宽高自动写入文件元数据，逆变换时据此裁剪输出尺寸。

| 选项 | 说明                                       | 默认值           |
| ---- | ------------------------------------------ | ---------------- |
| `-i` | 输入图片路径（必需）                       | —                |
| `-o` | 输出文件路径                               | 输入文件名 `.h5` |
| `-f` | 输出格式：`hdf5`，`tiff`=6通道float64 TIFF | `hdf5`           |

### `i` — 逆变换

从 `l` 命令生成的无损文件读取复数频谱，执行逆 FFT 后按原始尺寸裁剪输出。

| 选项 | 说明                                     | 默认值    |
| ---- | ---------------------------------------- | --------- |
| `-i` | 输入文件路径（.tiff / .tif / .h5，必需） | —         |
| `-o` | 输出 PNG 路径                            | `out.png` |

## 数据精度对比

| 模式             | 格式       | 位深                | 可逆 |
| ---------------- | ---------- | ------------------- | ---- |
| 有损 `f`         | PNG        | uint8               | ❌    |
| 有损 `f -t tiff` | TIFF       | float64（已归一化） | ❌    |
| 无损 `l`         | HDF5       | float64             | ✅    |
| 无损 `l -f tiff` | TIFF 6通道 | float64             | ✅    |

## 支持格式

**输入**：PNG / JPEG / GIF / BMP / TIFF / WebP

**输出**：PNG（8位）、TIFF（float64）、HDF5（float64）

## 依赖

- [go-dsp/fft](https://github.com/madelynnblue/go-dsp) — 快速傅里叶变换
- [scigolib/hdf5](https://github.com/scigolib/hdf5) — 纯 Go HDF5 读写
- [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) — 扩展图像格式解码

## 许可证

MIT
