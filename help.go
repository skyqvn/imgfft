package main

import "fmt"

func printHelp() {
	fmt.Print(`imgfft — 图像二维傅里叶变换工具

用法:  imgfft <command> [options]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
命令
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  v  view    有损正向 — 生成可视化频谱图，供人眼查看
  s  save    无损正向 — 保存完整复数频谱，可精确还原
  r  restore 逆变换   — 从无损文件还原原图
  h  help    帮助

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
v  view (有损正向变换)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  三通道分别 FFT 后对数归一化输出，不可逆变换。

  -i <path>    input  输入图片路径
  -o <prefix>  output 输出前缀           [默认: 输入文件名]
  -m <0|1>     mode   0=幅度+相位 1=实部+虚部  [默认: 0]
  -f <fmt>     format png=8位PNG tiff=float64 TIFF  [默认: png]

  输出: <prefix>_mag.png + <prefix>_phase.png 或 <prefix>_re.png + <prefix>_im.png

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
s  save (无损正向变换)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  保存完整复数频谱（float64），配合 r 命令可精确还原。
  原图宽高写入文件元数据，还原时自动裁剪。

  -i <path>  input  输入图片路径
  -o <path>  output 输出文件路径         [默认: 输入文件名.h5]
  -f <fmt>   format hdf5=HDF5复数 tiff=6通道float64  [默认: hdf5]

  TIFF: R_re G_re B_re R_im G_im B_im 交错存储

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
r  restore (逆变换)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  从 s 命令生成的无损文件读取复数频谱，IFFT 后还原。

  -i <path>  input  输入文件 (.tiff/.tif/.h5)
  -o <path>  output 输出 PNG             [默认: out.png]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
示例
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  imgfft v -i photo.png              幅度+相位 PNG
  imgfft v -i photo.png -m 1         实部+虚部 PNG
  imgfft v -i photo.png -f tiff      幅度+相位 TIFF
  imgfft s -i photo.png              无损 HDF5 → photo.h5
  imgfft s -i photo.png -f tiff      无损 TIFF → photo.tiff
  imgfft r -i photo.h5 -o out.png    从 HDF5 还原

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
精度
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  v (有损):  频谱经对数归一化，不可逆，仅可视化
    PNG   uint8
    TIFF  float64
  s (无损):  float64 完整精度，IFFT 可精确还原
    HDF5  float64 复数
    TIFF  float64 6通道（R_re G_re B_re R_im G_im B_im）

输入格式: PNG / JPEG / GIF / BMP / TIFF / WebP
`)
}
