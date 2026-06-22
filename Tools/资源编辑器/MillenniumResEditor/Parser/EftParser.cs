using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Text;
using MillenniumResEditor.Model;
using MillenniumResEditor.Parser;

namespace MillenniumResEditor.Parser
{
    /// <summary>
    /// 千年 EFT 特效文件解析器
    ///
    /// 格式规范（来自原版 Delphi 源码 A2Img.pas + AtzCls.pas）：
    ///
    /// EFT 文件与 ATZ 类似，但专门用于存储特效/技能动画
    ///
    /// TA2ImageLibEFD Header:
    ///   ImageLibEFD0 = 'EFD0' : 调色板模式（类似 ATZ0）
    ///   ImageLibEFD1 = 'EFD1' : RGB555 直接模式（类似 ATZ1）
    ///   ImageLibEFD2 = 'EFD2' : 加密模式（类似 ATZ3/4）
    ///
    /// TA2ImageEFD_FileHeader（$26 字节 = 38字节）：
    ///   IDent[2]     : 'EF' 标识
    ///   id(Word)     : 帧ID
    ///   px(SmallInt) : X偏移
    ///   py(SmallInt) : Y偏移
    ///   (剩余字节为扩展数据)
    ///
    /// 特效位置数据（AtzPosXY.sdb / TEffectPositionData）：
    ///   OverValue(Int32) : 超出值
    ///   Speed(Int32)     : 速度
    ///   Dir0X~Dir7X(Int32)×8 : 8个方向的X偏移
    ///   Dir0Y~Dir7Y(Int32)×8 : 8个方向的Y偏移
    ///
    /// 使用场景：
    ///   1. 从 eft.pgk 包中提取 .eft 特效文件
    ///   2. 解析并预览特效动画帧
    ///   3. 导出为 PNG 序列或 GIF 动图
    ///   4. 修改后重新打包导入游戏
    /// </summary>
    public class EftParser
    {
        // 格式标识常量
        public const string ID_EFD0 = "EFD0";
        public const string ID_EFD1 = "EFD1";
        public const string ID_EFD2 = "EFD2";

        /// <summary>解析出的所有帧</summary>
        public List<EftFrame> Frames { get; } = new List<EftFrame>();

        /// <summary>格式标识</summary>
        public string FormatId { get; private set; }

        /// <summary>透明色</summary>
        public int TransparentColor { get; private set; }

        /// <summary>调色板（EFD0 模式有效）</summary>
        public PalColor[] Palette { get; private set; }

        /// <summary>16字节密钥</summary>
        public byte[] RbyteKey { get; private set; } = new byte[16];

        /// <summary>从文件加载</summary>
        public void Load(string filePath)
        {
            using var fs = new FileStream(filePath, FileMode.Open, FileAccess.Read);
            using var br = new BinaryReader(fs);
            Parse(br);
        }

        /// <summary>从内存流加载</summary>
        public void LoadFromStream(Stream stream)
        {
            stream.Position = 0;
            using var br = new BinaryReader(stream);
            Parse(br);
        }

        private void Parse(BinaryReader br)
        {
            Frames.Clear();

            // 检测是否有 rbyte 前缀
            long startPos = br.BaseStream.Position;
            byte[] peek = br.ReadBytes(4);
            string peekIdent = Encoding.ASCII.GetString(peek);
            bool hasRbytePrefix = !IsValidIdent(peekIdent);

            br.BaseStream.Position = startPos;

            if (hasRbytePrefix)
            {
                for (int i = 0; i < 16; i++)
                    RbyteKey[i] = br.ReadByte();
            }
            else
            {
                Array.Clear(RbyteKey, 0, 16);
            }

            // 读取头部
            byte[] identBytes = br.ReadBytes(4);
            string ident = Encoding.ASCII.GetString(identBytes);

            if (!IsValidIdent(ident))
                throw new InvalidDataException($"未知的 EFT 格式标识：{ident}");

            int imageCount = br.ReadInt32();
            TransparentColor = br.ReadInt32();

            // 读取调色板
            Palette = new PalColor[256];
            for (int i = 0; i < 256; i++)
            {
                Palette[i] = new PalColor
                {
                    R = br.ReadByte(),
                    G = br.ReadByte(),
                    B = br.ReadByte()
                };
                br.ReadByte(); // Used
            }

            FormatId = ident;

            switch (ident)
            {
                case ID_EFD0:
                    ParsePaletteMode(br, imageCount, encrypted: false);
                    break;
                case ID_EFD1:
                    ParseRgb555Mode(br, imageCount, encrypted: false);
                    break;
                case ID_EFD2:
                    ParseRgb555Mode(br, imageCount, encrypted: true);
                    break;
                default:
                    throw new InvalidDataException($"未支持的 EFT 格式版本：{ident}");
            }
        }

        private bool IsValidIdent(string ident)
        {
            return ident == ID_EFD0 || ident == ID_EFD1 || ident == ID_EFD2;
        }

        private void ParsePaletteMode(BinaryReader br, int count, bool encrypted)
        {
            for (int n = 0; n < count; n++)
            {
                // EFD 使用 $26 字节头（38字节）
                ushort id = br.ReadUInt16();
                short px = br.ReadInt16();
                short py = br.ReadInt16();

                // 跳过剩余头字节（38 - 6 = 32字节）
                br.ReadBytes(32);

                int width = br.ReadInt32();
                int height = br.ReadInt32();

                if (encrypted)
                {
                    width = Rol32(width, 4) ^ RbyteKey[(n + 3) % 16];
                    height = Rol32(height, 2) ^ RbyteKey[(n + 5) % 16];
                }

                var frame = new EftFrame
                {
                    FrameId = id,
                    Width = (ushort)width,
                    Height = (ushort)height,
                    OffsetX = px,
                    OffsetY = py,
                    Format = EftFormat.Palette
                };

                int dataLen = width * height;
                byte[] indices = br.ReadBytes(dataLen);
                frame.PixelData = PaletteToRGB555(indices, Palette);
                Frames.Add(frame);
            }
        }

        private void ParseRgb555Mode(BinaryReader br, int count, bool encrypted)
        {
            for (int n = 0; n < count; n++)
            {
                ushort id = br.ReadUInt16();
                short px = br.ReadInt16();
                short py = br.ReadInt16();
                br.ReadBytes(32); // 跳过剩余头

                int width = br.ReadInt32();
                int height = br.ReadInt32();

                if (encrypted)
                {
                    width = Rol32(width, 4) ^ RbyteKey[(n + 3) % 16];
                    height = Rol32(height, 2) ^ RbyteKey[(n + 5) % 16];
                }

                var frame = new EftFrame
                {
                    FrameId = id,
                    Width = (ushort)width,
                    Height = (ushort)height,
                    OffsetX = px,
                    OffsetY = py,
                    Format = encrypted ? EftFormat.RGB555Encrypted : EftFormat.RGB555
                };

                frame.PixelData = br.ReadBytes(width * height * 2);
                Frames.Add(frame);
            }
        }

        private static byte[] PaletteToRGB555(byte[] indices, PalColor[] palette)
        {
            byte[] result = new byte[indices.Length * 2];
            for (int i = 0; i < indices.Length; i++)
            {
                PalColor c = palette[indices[i]];
                int r = c.R >> 3;
                if (c.R != 0 && r == 0) r = 1;
                int g = c.G >> 3;
                if (c.G != 0 && g == 0) g = 1;
                int b = c.B >> 3;
                if (c.B != 0 && b == 0) b = 1;

                ushort color555 = (ushort)((r << 10) | (g << 5) | b);
                result[i * 2] = (byte)(color555 & 0xFF);
                result[i * 2 + 1] = (byte)((color555 >> 8) & 0xFF);
            }
            return result;
        }

        private static int Rol32(int value, int bits)
        {
            uint v = (uint)value;
            return (int)((v << bits) | (v >> (32 - bits)));
        }

        /// <summary>将帧转换为 Bitmap</summary>
        public static Bitmap FrameToBitmap(EftFrame frame)
        {
            var bmp = new Bitmap(frame.Width, frame.Height, PixelFormat.Format32bppArgb);
            var rect = new Rectangle(0, 0, frame.Width, frame.Height);
            var bmpData = bmp.LockBits(rect, ImageLockMode.WriteOnly, PixelFormat.Format32bppArgb);

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

                            if (color555 == 0)
                            {
                                row[x * 4 + 3] = 0;
                            }
                            else
                            {
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

    /// <summary>
    /// EFT 单帧数据
    /// </summary>
    public class EftFrame
    {
        /// <summary>帧ID</summary>
        public ushort FrameId { get; set; }

        /// <summary>宽度</summary>
        public ushort Width { get; set; }

        /// <summary>高度</summary>
        public ushort Height { get; set; }

        /// <summary>X偏移</summary>
        public short OffsetX { get; set; }

        /// <summary>Y偏移</summary>
        public short OffsetY { get; set; }

        /// <summary>RGB555 像素数据</summary>
        public byte[] PixelData { get; set; }

        /// <summary>格式</summary>
        public EftFormat Format { get; set; }
    }

    /// <summary>
    /// EFT 格式枚举
    /// </summary>
    public enum EftFormat
    {
        Palette,
        RGB555,
        RGB555Encrypted
    }
}
