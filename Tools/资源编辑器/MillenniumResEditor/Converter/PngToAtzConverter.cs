using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Text;
using MillenniumResEditor.Model;
using MillenniumResEditor.Parser;

namespace MillenniumResEditor.Converter
{
    /// <summary>
    /// PNG 图片序列转 ATZ 文件转换器（逆向转换）
    ///
    /// 支持将 PNG/BMP 图片序列打包回 ATZ 格式，用于修改后导入游戏
    ///
    /// 输出格式：
    ///   ATZ1: RGB555 直接模式（最通用，兼容性最好）
    ///   ATZ0: 调色板模式（需要提供调色板或自动生成）
    ///   ATZ5: PNG 压缩模式（保留原始PNG数据）
    ///
    /// 使用场景：
    ///   1. 从 ATZ 导出 PNG → 修改图片 → 重新打包为 ATZ → 替换游戏资源
    ///   2. 自定义角色/装备制作 → 打包为 ATZ → 导入游戏测试
    /// </summary>
    public class PngToAtzConverter
    {
        /// <summary>帧列表</summary>
        public List<AtzFrame> Frames { get; } = new List<AtzFrame>();

        /// <summary>输出格式</summary>
        public AtzFormat OutputFormat { get; set; } = AtzFormat.RGB555;

        /// <summary>透明色值（RGB555）</summary>
        public ushort TransparentColor { get; set; } = 0;

        /// <summary>调色板（仅 ATZ0 模式需要）</summary>
        public PalColor[] Palette { get; set; }

        /// <summary>16字节密钥头（用于 ATZ3/ATZ4 加密模式）</summary>
        public byte[] RbyteKey { get; set; } = new byte[16];

        /// <summary>
        /// 从 PNG 文件列表加载帧
        /// </summary>
        /// <param name="pngFiles">PNG 文件路径列表（按顺序）</param>
        /// <param name="autoDetectOffset">是否自动检测偏移量（通过透明边距）</param>
        public void LoadFromPngFiles(string[] pngFiles, bool autoDetectOffset = true)
        {
            Frames.Clear();

            foreach (string filePath in pngFiles)
            {
                if (!File.Exists(filePath)) continue;

                using var bmp = new Bitmap(filePath);
                var frame = BitmapToFrame(bmp, autoDetectOffset);
                Frames.Add(frame);
            }
        }

        /// <summary>
        /// 从目录批量加载 PNG（按文件名排序）
        /// </summary>
        /// <param name="directory">目录路径</param>
        /// <param name="searchPattern">搜索模式（如 "frame_*.png"）</param>
        /// <param name="autoDetectOffset">是否自动检测偏移</param>
        public void LoadFromDirectory(string directory, string searchPattern = "*.png", bool autoDetectOffset = true)
        {
            if (!Directory.Exists(directory)) return;

            string[] files = Directory.GetFiles(directory, searchPattern);
            Array.Sort(files); // 按文件名排序
            LoadFromPngFiles(files, autoDetectOffset);
        }

        /// <summary>
        /// 添加单张图片为帧
        /// </summary>
        public void AddFrame(Bitmap bmp, bool autoDetectOffset = true)
        {
            var frame = BitmapToFrame(bmp, autoDetectOffset);
            Frames.Add(frame);
        }

        /// <summary>
        /// 将 Bitmap 转换为 AtzFrame（ARGB32 → RGB555）
        /// </summary>
        private AtzFrame BitmapToFrame(Bitmap bmp, bool autoDetectOffset)
        {
            var frame = new AtzFrame
            {
                Width = (ushort)bmp.Width,
                Height = (ushort)bmp.Height,
                Format = OutputFormat
            };

            // 自动检测偏移（找到非透明区域的中心点）
            if (autoDetectOffset)
            {
                var (ox, oy) = DetectOffset(bmp);
                frame.OffsetX = (short)ox;
                frame.OffsetY = (short)oy;
            }

            // 转换像素数据
            if (OutputFormat == AtzFormat.Png)
            {
                // ATZ5: 保存原始 PNG 数据
                using var ms = new MemoryStream();
                bmp.Save(ms, ImageFormat.Png);
                frame.PngData = ms.ToArray();
            }
            else
            {
                // ATZ0/ATZ1: 转换为 RGB555
                frame.PixelData = BitmapToRgb555(bmp);
            }

            return frame;
        }

        /// <summary>
        /// 检测图像偏移（基于透明边距，计算到画布中心的偏移）
        /// </summary>
        private (int offsetX, int offsetY) DetectOffset(Bitmap bmp)
        {
            int minX = bmp.Width, minY = bmp.Height, maxX = 0, maxY = 0;
            bool hasContent = false;

            var rect = new Rectangle(0, 0, bmp.Width, bmp.Height);
            var bmpData = bmp.LockBits(rect, ImageLockMode.ReadOnly, PixelFormat.Format32bppArgb);

            try
            {
                unsafe
                {
                    byte* scan0 = (byte*)bmpData.Scan0;
                    int stride = bmpData.Stride;

                    for (int y = 0; y < bmp.Height; y++)
                    {
                        byte* row = scan0 + y * stride;
                        for (int x = 0; x < bmp.Width; x++)
                        {
                            if (row[x * 4 + 3] > 0) // Alpha > 0
                            {
                                if (x < minX) minX = x;
                                if (x > maxX) maxX = x;
                                if (y < minY) minY = y;
                                if (y > maxY) maxY = y;
                                hasContent = true;
                            }
                        }
                    }
                }
            }
            finally
            {
                bmp.UnlockBits(bmpData);
            }

            if (!hasContent) return (0, 0);

            // 计算内容中心相对于图像中心的偏移
            int contentCenterX = (minX + maxX) / 2;
            int contentCenterY = (minY + maxY) / 2;
            int imageCenterX = bmp.Width / 2;
            int imageCenterY = bmp.Height / 2;

            return (contentCenterX - imageCenterX, contentCenterY - imageCenterY);
        }

        /// <summary>
        /// Bitmap (ARGB32) → RGB555 字节数组
        /// </summary>
        private static byte[] BitmapToRgb555(Bitmap bmp)
        {
            int w = bmp.Width;
            int h = bmp.Height;
            byte[] result = new byte[w * h * 2];

            var rect = new Rectangle(0, 0, w, h);
            var bmpData = bmp.LockBits(rect, ImageLockMode.ReadOnly, PixelFormat.Format32bppArgb);

            try
            {
                unsafe
                {
                    byte* src = (byte*)bmpData.Scan0;
                    int srcStride = bmpData.Stride;

                    for (int y = 0; y < h; y++)
                    {
                        byte* srcRow = src + y * srcStride;
                        for (int x = 0; x < w; x++)
                        {
                            byte b = srcRow[x * 4 + 0];
                            byte g = srcRow[x * 4 + 1];
                            byte r = srcRow[x * 4 + 2];
                            byte a = srcRow[x * 4 + 3];

                            int idx = (y * w + x) * 2;

                            if (a == 0)
                            {
                                // 完全透明 → 写入透明色(0)
                                result[idx] = 0;
                                result[idx + 1] = 0;
                            }
                            else
                            {
                                // RGB888 → RGB555
                                int r5 = r >> 3;
                                int g5 = g >> 3;
                                int b5 = b >> 3;
                                ushort color555 = (ushort)((r5 << 10) | (g5 << 5) | b5);

                                result[idx] = (byte)(color555 & 0xFF);
                                result[idx + 1] = (byte)((color555 >> 8) & 0xFF);
                            }
                        }
                    }
                }
            }
            finally
            {
                bmp.UnlockBits(bmpData);
            }

            return result;
        }

        /// <summary>
        /// 保存为 ATZ 文件
        /// </summary>
        /// <param name="outputPath">输出文件路径</param>
        public void Save(string outputPath)
        {
            if (Frames.Count == 0)
                throw new InvalidOperationException("没有可保存的帧数据");

            using var fs = new FileStream(outputPath, FileMode.Create, FileAccess.Write);
            using var bw = new BinaryWriter(fs);

            // 1. 写入 16 字节 rbyte 密钥头（用于 ATZ3/ATZ4）
            bw.Write(RbyteKey);

            // 2. 写入 TA2ImageLibHeader
            string formatId = GetFormatId();
            bw.Write(Encoding.ASCII.GetBytes(formatId)); // Ident[4]
            bw.Write(Frames.Count);                      // ImageCount
            bw.Write(TransparentColor);                  // TransparentColor

            // 3. 写入调色板（256 × 4字节：R+G+B+Used）
            if (OutputFormat == AtzFormat.Palette || OutputFormat == AtzFormat.PaletteEncrypted)
            {
                if (Palette == null)
                    Palette = GenerateDefaultPalette();

                for (int i = 0; i < 256; i++)
                {
                    bw.Write(Palette[i].R);
                    bw.Write(Palette[i].G);
                    bw.Write(Palette[i].B);
                    bw.Write((byte)1); // Used = 1
                }
            }
            else
            {
                // 非调色板模式写入零填充的调色板区域
                bw.Write(new byte[256 * 4]);
            }

            // 4. 写入每帧数据
            for (int n = 0; n < Frames.Count; n++)
            {
                var frame = Frames[n];

                switch (OutputFormat)
                {
                    case AtzFormat.Palette:
                        WritePaletteFrame(bw, frame, n, encrypted: false);
                        break;
                    case AtzFormat.RGB555:
                        WriteRgb555Frame(bw, frame, n, encrypted: false);
                        break;
                    case AtzFormat.PaletteEncrypted:
                        WritePaletteFrame(bw, frame, n, encrypted: true);
                        break;
                    case AtzFormat.RGB555Encrypted:
                        WriteRgb555Frame(bw, frame, n, encrypted: true);
                        break;
                    case AtzFormat.Png:
                        WritePngFrame(bw, frame);
                        break;
                }
            }
        }

        /// <summary>获取当前格式的标识字符串</summary>
        private string GetFormatId()
        {
            return OutputFormat switch
            {
                AtzFormat.Palette => "ATZ0",
                AtzFormat.RGB555 => "ATZ1",
                AtzFormat.PaletteEncrypted => "ATZ3",
                AtzFormat.RGB555Encrypted => "ATZ4",
                AtzFormat.Png => "ATZ5",
                _ => "ATZ1"
            };
        }

        /// <summary>写调色板模式帧（ATZ0/ATZ3）</summary>
        private void WritePaletteFrame(BinaryWriter bw, AtzFrame frame, int index, bool encrypted)
        {
            int w = frame.Width;
            int h = frame.Height;

            if (encrypted)
            {
                // ATZ3: 宽高加密
                int ew = Ror32(w, 4) ^ RbyteKey[(index + 3) % 16];
                int eh = Ror32(h, 2) ^ RbyteKey[(index + 5) % 16];
                bw.Write(ew);
                bw.Write(eh);
            }
            else
            {
                bw.Write(w);
                bw.Write(h);
            }

            bw.Write(frame.OffsetX);
            bw.Write(frame.OffsetY);

            // 将 RGB555 转回调色板索引（使用最近邻颜色匹配）
            byte[] indices = Rgb555ToPaletteIndices(frame.PixelData, Palette);
            bw.Write(indices);
        }

        /// <summary>写 RGB555 模式帧（ATZ1/ATZ4）</summary>
        private void WriteRgb555Frame(BinaryWriter bw, AtzFrame frame, int index, bool encrypted)
        {
            int w = frame.Width;
            int h = frame.Height;

            if (encrypted)
            {
                int ew = Ror32(w, 4) ^ RbyteKey[(index + 3) % 16];
                int eh = Ror32(h, 2) ^ RbyteKey[(index + 5) % 16];
                bw.Write(ew);
                bw.Write(eh);
            }
            else
            {
                bw.Write(w);
                bw.Write(h);
            }

            bw.Write(frame.OffsetX);
            bw.Write(frame.OffsetY);

            bw.Write(frame.PixelData);
        }

        /// <summary>写 PNG 模式帧（ATZ5）</summary>
        private void WritePngFrame(BinaryWriter bw, AtzFrame frame)
        {
            bw.Write(frame.Width);
            bw.Write(frame.Height);
            bw.Write(frame.OffsetX);
            bw.Write(frame.OffsetY);
            bw.Write(frame.PngData?.Length ?? 0); // filesize
            if (frame.PngData != null)
                bw.Write(frame.PngData);
        }

        /// <summary>
        /// RGB555 → 调色板索引（最近邻匹配）
        /// </summary>
        private static byte[] Rgb555ToPaletteIndices(byte[] rgb555Data, PalColor[] palette)
        {
            if (palette == null || rgb555Data == null) return Array.Empty<byte>();

            int pixelCount = rgb555Data.Length / 2;
            byte[] indices = new byte[pixelCount];

            for (int i = 0; i < pixelCount; i++)
            {
                ushort c555 = (ushort)(rgb555Data[i * 2] | (rgb555Data[i * 2 + 1] << 8));

                if (c555 == 0)
                {
                    indices[i] = 0; // 透明色
                    continue;
                }

                int r = (c555 >> 10) & 0x1F;
                int g = (c555 >> 5) & 0x1F;
                int b = c555 & 0x1F;

                // 最近邻搜索最佳匹配调色板索引
                int bestIdx = 0;
                int minDist = int.MaxValue;

                for (int j = 0; j < 256; j++)
                {
                    int pr = palette[j].R >> 3;
                    int pg = palette[j].G >> 3;
                    int pb = palette[j].B >> 3;

                    int dist = (r - pr) * (r - pr) +
                               (g - pg) * (g - pg) +
                               (b - pb) * (b - pb);

                    if (dist < minDist)
                    {
                        minDist = dist;
                        bestIdx = j;
                    }
                }

                indices[i] = (byte)bestIdx;
            }

            return indices;
        }

        /// <summary>生成默认 256 色调色板（均匀分布）</summary>
        private static PalColor[] GenerateDefaultPalette()
        {
            var pal = new PalColor[256];

            // 0 号色为透明色（黑色）
            pal[0] = new PalColor { R = 0, G = 0, B = 0 };

            // 其余 255 色均匀分布
            for (int i = 1; i < 256; i++)
            {
                // 简单的 RGB 立方体分布
                int step = (int)Math.Ceiling(Math.Pow(255, 1.0 / 3.0));
                int ri = (i - 1) / (step * step);
                int gi = ((i - 1) / step) % step;
                int bi = (i - 1) % step;

                pal[i] = new PalColor
                {
                    R = (byte)(ri * 255 / Math.Max(step - 1, 1)),
                    G = (byte)(gi * 255 / Math.Max(step - 1, 1)),
                    B = (byte)(bi * 255 / Math.Max(step - 1, 1))
                };
            }

            return pal;
        }

        /// <summary>32位循环右移（ROR）</summary>
        private static int Ror32(int value, int bits)
        {
            uint v = (uint)value;
            return (int)((v >> bits) | (v << (32 - bits)));
        }
    }
}
