package main

import (
	"fmt"
)

func printHelp() {
	fmt.Print(`imgfft - 图像二维傅里叶变换工具

用法:
  imgfft <命令> [选项]

命令:
  f    有损正向变换（输出可视化频谱图）
  l    无损正向变换（保存完整复数频谱，可精确逆变换还原）
  i    逆变换（从无损文件还原原图）
  h    显示本帮助

------------------------------------------------------------
f 选项（有损正向变换）:
  -i <路径>    输入图片（必需）
  -o <前缀>    输出前缀（默认=输入文件名）
  -m <模式>    输出模式: 0=幅度+相位（默认）, 1=实部+虚部
  -t <格式>    输出格式: png=8位PNG, tiff=float64 TIFF（默认png）

  说明: 三通道分别变换后合成到一张 RGB 图，已做归一化，适合
        直接查看频谱形态，但无法逆变换精确还原。
        模式0 输出 <前缀>_mag / <前缀>_phase。
        模式1 输出 <前缀>_re / <前缀>_im。
        输出尺寸 = FFT 尺寸（大于等于原图的下一个 2 的幂）。

------------------------------------------------------------
l 选项（无损正向变换）:
  -i <路径>    输入图片（必需）
  -o <路径>    输出文件（默认=输入文件名.tiff 或 .h5）
  -f <格式>    输出格式: tiff=6通道float64 TIFF（默认）, hdf5=HDF5

  说明: 保存完整复数频谱（float64），不做任何归一化/量化，
        配合 i 命令可精确还原原图（误差仅限浮点舍入）。
        原图宽高写入 TIFF ImageDescription 标签或 HDF5 orig_size 属性，
        逆变换时自动据此裁剪输出。
        TIFF 为 6 通道交错: R_re G_re B_re R_im G_im B_im。
        FFT 尺寸自动取不小于 max(宽,高) 的下一个 2 的幂。

------------------------------------------------------------
i 选项（逆变换）:
  -i <路径>    输入文件（.tiff / .tif / .h5，必需）
  -o <路径>    输出 PNG（默认 out.png）

  说明: 从 l 命令生成的无损文件读取复数频谱，做逆 FFT 后
        按原图尺寸裁剪输出 PNG。
        TIFF 自动识别 32/64 位，兼容旧版本文件。

------------------------------------------------------------
示例:
  imgfft f -i photo.png              有损正向，输出 photo_mag.png 和 photo_phase.png
  imgfft f -i photo.png -m 1        输出实部/虚部图
  imgfft f -i photo.png -t tiff      输出 float64 TIFF（有损，已归一化）
  imgfft l -i photo.png              无损正向，输出 photo.tiff
  imgfft l -i photo.png -f hdf5      无损正向，输出 photo.h5
  imgfft i -i photo.tiff             逆变换，输出 out.png
  imgfft i -i photo.h5 -o r.png      从 HDF5 逆变换

数据精度说明:
  有损模式: 频谱经过归一化，信息不可逆，仅用于可视化。
  无损模式: float64 完整精度，IFFT 后可精确还原原图。
    PNG:  每通道 uint8，有损
    TIFF: 每通道 float32/float64，6 通道时无损
    HDF5: float64 复数，无损

支持格式:
  输入图片: PNG / JPEG / GIF（标准库）。
`)
}
