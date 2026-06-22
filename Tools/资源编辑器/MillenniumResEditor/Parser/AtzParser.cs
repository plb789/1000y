using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Text;
using MillenniumResEditor.Model;
using MillenniumResEditor.Utils;

namespace MillenniumResEditor.Parser
{
    /// <summary>
    /// 千年 .atz 文件解析器（TA2ImageLib 二进制格式）
    ///
    /// 格式规范（来自原版 Delphi 源码 A2Img.pas + pgktools/Unit1.pas）：
    ///
    /// 文件整体布局：
    ///   [0-15]   : 16字节 rbyte 密钥（用于 ID3/ID4 格式的宽高解密）
    ///   [16-...]  : TA2ImageLibHeader（16字节 Ident + ImageCount + TransparentColor + Palette）
    ///   之后为每帧的 FileHeader + PixelData
    ///
    /// TA2ImageLibHeader（1032字节）：
    ///   Ident[4]          : 'ATZ0'/'ATZ1'/'ATZ3'/'ATZ4'/'ATZ5'
    ///   ImageCount(Int32) : 图像数量
    ///   TransparentColor  : 透明色
    ///   Palette[256*4]    : TImgLibPalette（仅 ATZ0/ATZ3 使用，每项 RGB+Used）
    ///
    /// 5种格式版本：
    ///   ATZ0：调色板模式（每像素1字节索引，通过 Palette 转换为 RGB555）
    ///   ATZ1：直接 RGB555（每像素2字节）
    ///   ATZ3：加密调色板模式（宽高需 rol 4/2 + xor rbyte 解密）
    ///   ATZ4：加密 RGB555 模式（宽高需解密）
    ///   ATZ5：PNG 压缩模式（每帧内嵌完整 PNG 数据）
    ///
    /// TA2ImageFileHeader（16字节，ATZ0/1/3/4）：
    ///   Width(Int32)  Height(Int32)  px(Int32)  py(Int32)
    ///
    /// TA2ImageFileHeaderForPng（20字节，ATZ5）：
    ///   Width(Int32)  Height(Int32)  px(Int32)  py(Int32)  filesize(Int32)
    /// </summary>
    public class AtzParser
    {
        // 格式标识常量（来自 A2Img.pas）
        public const string ID_ATZ0 = "ATZ0"; // 调色板模式
        public const string ID_ATZ1 = "ATZ1"; // RGB555 直接模式
        public const string ID_ATZ3 = "ATZ3"; // 加密调色板模式
        public const string ID_ATZ4 = "ATZ4"; // 加密 RGB555 模式
        public const string ID_ATZ5 = "ATZ5"; // PNG 压缩模式

        /// <summary>解析出的所有帧</summary>
        public List<AtzFrame> Frames { get; } = new List<AtzFrame>();

        /// <summary>格式标识</summary>
        public string FormatId { get; private set; }

        /// <summary>透明色</summary>
        public int TransparentColor { get; private set; }

        /// <summary>调色板（ATZ0/ATZ3 模式有效）</summary>
        public PalColor[] Palette { get; private set; }

        /// <summary>16字节密钥（用于 ATZ3/ATZ4 宽高解密）</summary>
        public byte[] RbyteKey { get; private set; } = new byte[16];

        /// <summary>从文件加载 .atz</summary>
        public void Load(string filePath)
        {
            using var fs = new FileStream(filePath, FileMode.Open, FileAccess.Read);
            using var br = new BinaryReader(fs);
            Parse(br);
        }

        /// <summary>从内存流加载（用于从 PGK 包中提取后直接解析）</summary>
        public void LoadFromStream(Stream stream)
        {
            stream.Position = 0;
            using var br = new BinaryReader(stream);
            Parse(br);
        }

        private void Parse(BinaryReader br)
        {
            Frames.Clear();

            // 检测格式：peek 前4字节判断是否有 16字节 rbyte 前缀
            long startPos = br.BaseStream.Position;
            byte[] peek = br.ReadBytes(4);
            string peekIdent = Encoding.ASCII.GetString(peek);
            bool hasRbytePrefix = !IsValidIdent(peekIdent);

            // 回到起始位置
            br.BaseStream.Position = startPos;

            if (hasRbytePrefix)
            {
                // 格式A：有 16 字节 rbyte 前缀（原始 .atz 文件）
                for (int i = 0; i < 16; i++)
                    RbyteKey[i] = br.ReadByte();
            }
            else
            {
                // 格式B：无前缀，头部直接开始（PGK 包内数据）
                // 用零填充 RbyteKey
                Array.Clear(RbyteKey, 0, 16);
            }

            // 读取 TA2ImageLibHeader
            byte[] identBytes = br.ReadBytes(4);
            string ident = Encoding.ASCII.GetString(identBytes);

            if (!IsValidIdent(ident))
                throw new InvalidDataException($"未知的 .atz 格式标识：{ident}");

            int imageCount = br.ReadInt32();
            TransparentColor = br.ReadInt32();

            // 读取调色板（256 * TImgLibColor，每个4字节：R+G+B+Used）
            Palette = new PalColor[256];
            for (int i = 0; i < 256; i++)
            {
                Palette[i] = new PalColor
                {
                    R = br.ReadByte(),
                    G = br.ReadByte(),
                    B = br.ReadByte()
                };
                br.ReadByte(); // Used 字段
            }

            FormatId = ident;

            // 3. 根据格式版本解析每帧
            switch (ident)
            {
                case ID_ATZ0:
                    ParsePaletteMode(br, imageCount, encrypted: false);
                    break;
                case ID_ATZ1:
                    ParseRGB555Mode(br, imageCount, encrypted: false);
                    break;
                case ID_ATZ3:
                    ParsePaletteMode(br, imageCount, encrypted: true);
                    break;
                case ID_ATZ4:
                    ParseRGB555Mode(br, imageCount, encrypted: true);
                    break;
                case ID_ATZ5:
                    ParsePngMode(br, imageCount);
                    break;
                default:
                    throw new InvalidDataException($"未支持的 .atz 格式版本：{ident}");
            }
        }

        /// <summary>检查是否是有效的 ATZ 格式标识</summary>
        private bool IsValidIdent(string ident)
        {
            return ident == ID_ATZ0 || ident == ID_ATZ1 ||
                   ident == ID_ATZ3 || ident == ID_ATZ4 || ident == ID_ATZ5;
        }

        /// <summary>
        /// 解析调色板模式（ATZ0/ATZ3）
        /// 每帧：FileHeader(16) + 像素索引数据(Width*Height*1字节)
        /// </summary>
        private void ParsePaletteMode(BinaryReader br, int count, bool encrypted)
        {
            for (int n = 0; n < count; n++)
            {
                // 检查是否还有足够的数据读取帧头（16字节）
                if (br.BaseStream.Position + 16 > br.BaseStream.Length)
                {
                    Debug.WriteLine($"[警告] 帧索引 {n}: 数据不足，提前结束");
                    break;
                }

                int width = br.ReadInt32();
                int height = br.ReadInt32();
                int px = br.ReadInt32();
                int py = br.ReadInt32();

                if (encrypted)
                {
                    // ATZ3: w = (rol(w,4)) xor rbyte[(n+3)%16]
                    //       h = (rol(h,2)) xor rbyte[(n+5)%16]
                    width = Rol32(width, 4) ^ RbyteKey[(n + 3) % 16];
                    height = Rol32(height, 2) ^ RbyteKey[(n + 5) % 16];
                }

                // 边界检查：防止异常值导致内存溢出
                if (width <= 0 || height <= 0 || width > 4096 || height > 4096)
                {
                    Debug.WriteLine($"[警告] 帧索引 {n}: 异常尺寸 {width}x{height}，跳过此帧");
                    break;
                }

                long dataLen = (long)width * height;

                // 检查数据大小是否合理（单帧不超过20MB，调色板模式1字节/像素）
                if (dataLen > 20 * 1024 * 1024)
                {
                    Debug.WriteLine($"[警告] 帧索引 {n}: 数据过大 ({dataLen / 1024}KB)，跳过");
                    break;
                }

                // 检查剩余数据是否足够
                if (br.BaseStream.Position + dataLen > br.BaseStream.Length)
                {
                    long remaining = br.BaseStream.Length - br.BaseStream.Position;
                    Debug.WriteLine($"[警告] 帧索引 {n}: 需要 {dataLen} 字节，但仅剩 {remaining} 字节");
                    break;
                }

                var frame = new AtzFrame
                {
                    Width = (ushort)width,
                    Height = (ushort)height,
                    OffsetX = (short)px,
                    OffsetY = (short)py,
                    Format = encrypted ? AtzFormat.PaletteEncrypted : AtzFormat.Palette
                };

                // 像素索引数据（每像素1字节，通过调色板转换为颜色）
                byte[] indices = br.ReadBytes((int)dataLen);

                // 立即转换为 RGB555 像素数据
                frame.PixelData = PaletteToRGB555(indices, Palette);
                Frames.Add(frame);
            }
        }

        /// <summary>
        /// 解析 RGB555 直接模式（ATZ1/ATZ4）
        /// 每帧：FileHeader(16) + 像素数据(Width*Height*2字节)
        /// </summary>
        private void ParseRGB555Mode(BinaryReader br, int count, bool encrypted)
        {
            for (int n = 0; n < count; n++)
            {
                // 检查是否还有足够的数据读取帧头（16字节）
                if (br.BaseStream.Position + 16 > br.BaseStream.Length)
                {
                    Debug.WriteLine($"[警告] 帧索引 {n}: 数据不足，提前结束");
                    break;
                }

                int width = br.ReadInt32();
                int height = br.ReadInt32();
                int px = br.ReadInt32();
                int py = br.ReadInt32();

                if (encrypted)
                {
                    width = Rol32(width, 4) ^ RbyteKey[(n + 3) % 16];
                    height = Rol32(height, 2) ^ RbyteKey[(n + 5) % 16];
                }

                // 边界检查：防止异常值导致内存溢出
                if (width <= 0 || height <= 0 || width > 4096 || height > 4096)
                {
                    Debug.WriteLine($"[警告] 帧索引 {n}: 异常尺寸 {width}x{height}，跳过此帧");
                    // 尝试跳过（无法确定准确大小时只能停止）
                    break;
                }

                long dataSize = (long)width * height * 2;
                
                // 检查数据大小是否合理（单帧不超过10MB）
                if (dataSize > 10 * 1024 * 1024)
                {
                    Debug.WriteLine($"[警告] 帧索引 {n}: 数据过大 ({dataSize / 1024}KB)，跳过");
                    break;
                }

                // 检查剩余数据是否足够
                if (br.BaseStream.Position + dataSize > br.BaseStream.Length)
                {
                    long remaining = br.BaseStream.Length - br.BaseStream.Position;
                    Debug.WriteLine($"[警告] 帧索引 {n}: 需要 {dataSize} 字节，但仅剩 {remaining} 字节");
                    break;
                }

                var frame = new AtzFrame
                {
                    Width = (ushort)width,
                    Height = (ushort)height,
                    OffsetX = (short)px,
                    OffsetY = (short)py,
                    Format = encrypted ? AtzFormat.RGB555Encrypted : AtzFormat.RGB555
                };

                // RGB555 像素数据（每像素2字节）
                frame.PixelData = br.ReadBytes((int)dataSize);
                Frames.Add(frame);
            }
        }

        /// <summary>
        /// 解析 PNG 压缩模式（ATZ5）
        /// 每帧：FileHeaderForPng(20) + PNG 数据(filesize 字节)
        /// </summary>
        private void ParsePngMode(BinaryReader br, int count)
        {
            for (int n = 0; n < count; n++)
            {
                // 检查是否还有足够的数据读取帧头（20字节）
                if (br.BaseStream.Position + 20 > br.BaseStream.Length)
                {
                    Debug.WriteLine($"[警告] 帧索引 {n}: 数据不足，提前结束");
                    break;
                }

                int width = br.ReadInt32();
                int height = br.ReadInt32();
                int px = br.ReadInt32();
                int py = br.ReadInt32();
                int fileSize = br.ReadInt32();

                // 边界检查
                if (width <= 0 || height <= 0 || width > 4096 || height > 4096)
                {
                    Debug.WriteLine($"[警告] 帧索引 {n}: 异常尺寸 {width}x{height}，跳过此帧");
                    break;
                }

                // 检查文件大小是否合理（单帧PNG不超过20MB）
                if (fileSize <= 0 || fileSize > 20 * 1024 * 1024)
                {
                    Debug.WriteLine($"[警告] 帧索引 {n}: PNG大小异常 ({fileSize}字节)，跳过");
                    break;
                }

                // 检查剩余数据是否足够
                if (br.BaseStream.Position + fileSize > br.BaseStream.Length)
                {
                    long remaining = br.BaseStream.Length - br.BaseStream.Position;
                    Debug.WriteLine($"[警告] 帧索引 {n}: 需要 {fileSize} 字节，但仅剩 {remaining} 字节");
                    break;
                }

                var frame = new AtzFrame
                {
                    Width = (ushort)width,
                    Height = (ushort)height,
                    OffsetX = (short)px,
                    OffsetY = (short)py,
                    Format = AtzFormat.Png
                };

                // 直接保存 PNG 数据，转换时直接写入文件
                frame.PngData = br.ReadBytes(fileSize);
                Frames.Add(frame);
            }
        }

        /// <summary>
        /// 调色板索引转 RGB555 像素数据
        /// 对应原版 AnsPaletteDataToAns2Image
        /// </summary>
        private static byte[] PaletteToRGB555(byte[] indices, PalColor[] palette)
        {
            byte[] result = new byte[indices.Length * 2];
            for (int i = 0; i < indices.Length; i++)
            {
                PalColor c = palette[indices[i]];
                // RGB888 → RGB555（r>>3，非零值若结果为0则设为1）
                int r = c.R >> 3;
                if (c.R != 0 && r == 0) r = 1;
                int g = c.G >> 3;
                if (c.G != 0 && g == 0) g = 1;
                int b = c.B >> 3;
                if (c.B != 0 && b == 0) b = 1;

                // RGB555: rrrrrgggggbbbbbb（bit15=0, r=10-14, g=5-9, b=0-4）
                ushort color555 = (ushort)((r << 10) | (g << 5) | b);
                result[i * 2] = (byte)(color555 & 0xFF);
                result[i * 2 + 1] = (byte)((color555 >> 8) & 0xFF);
            }
            return result;
        }

        /// <summary>32位循环左移</summary>
        private static int Rol32(int value, int bits)
        {
            uint v = (uint)value;
            return (int)((v << bits) | (v >> (32 - bits)));
        }

        /// <summary>
        /// 将一帧转换为 System.Drawing.Bitmap
        /// RGB555 → ARGB32，透明色为 0
        /// </summary>
        public static System.Drawing.Bitmap FrameToBitmap(AtzFrame frame)
        {
            // PNG 格式直接从 PngData 加载
            if (frame.Format == AtzFormat.Png && frame.PngData != null && frame.PngData.Length > 0)
            {
                using var ms = new MemoryStream(frame.PngData);
                return new System.Drawing.Bitmap(ms);
            }

            var bmp = new System.Drawing.Bitmap(frame.Width, frame.Height,
                System.Drawing.Imaging.PixelFormat.Format32bppArgb);

            var rect = new System.Drawing.Rectangle(0, 0, frame.Width, frame.Height);
            var bmpData = bmp.LockBits(rect, System.Drawing.Imaging.ImageLockMode.WriteOnly,
                System.Drawing.Imaging.PixelFormat.Format32bppArgb);

            try
            {
                unsafe
                {
                    byte* dst = (byte*)bmpData.Scan0;
                    int stride = bmpData.Stride;

                    for (int y = 0; y < frame.Height; y++)
                    {
                        byte* row = dst + y * stride;
                        for (int x = 0; x < frame.Width; x++)
                        {
                            int srcIdx = (y * frame.Width + x) * 2;
                            if (srcIdx + 1 >= frame.PixelData.Length) break;

                            ushort color555 = (ushort)(frame.PixelData[srcIdx] | (frame.PixelData[srcIdx + 1] << 8));

                            // 透明色：0
                            if (color555 == 0)
                            {
                                row[x * 4 + 0] = 0;
                                row[x * 4 + 1] = 0;
                                row[x * 4 + 2] = 0;
                                row[x * 4 + 3] = 0;
                            }
                            else
                            {
                                // RGB555 → RGB888
                                int r = (color555 >> 10) & 0x1F;
                                int g = (color555 >> 5) & 0x1F;
                                int b = color555 & 0x1F;

                                row[x * 4 + 0] = (byte)((b << 3) | (b >> 2));
                                row[x * 4 + 1] = (byte)((g << 3) | (g >> 2));
                                row[x * 4 + 2] = (byte)((r << 3) | (r >> 2));
                                row[x * 4 + 3] = 255;
                            }
                        }
                    }
                }
            }
            finally
            {
                bmp.UnlockBits(bmpData);
            }

            return bmp;
        }
    }
}
