using System.Collections.Generic;
using System.IO;
using MillenniumResEditor.Model;
using MillenniumResEditor.Utils;

namespace MillenniumResEditor.Parser
{
    /// <summary>
    /// 千年 .spr 帧动画解析
    /// 32B头 + 帧数(2) + 逐帧数据
    /// </summary>
    public class SprParser
    {
        public List<SprFrame> Frames { get; } = new List<SprFrame>();
        private byte[] _fileHeader;

        public void Load(string filePath)
        {
            using var fs = new FileStream(filePath, FileMode.Open, FileAccess.Read);
            using var br = new BinaryReader(fs);

            // 跳过32字节文件头
            _fileHeader = br.ReadBytes(32);

            // 读取总帧数
            ushort frameCount = BinaryUtil.ReadUInt16LE(br);
            Frames.Clear();

            for (int i = 0; i < frameCount; i++)
            {
                var frame = new SprFrame
                {
                    Width = BinaryUtil.ReadUInt16LE(br),
                    Height = BinaryUtil.ReadUInt16LE(br),
                    OffsetX = BinaryUtil.ReadUInt16LE(br),
                    OffsetY = BinaryUtil.ReadUInt16LE(br)
                };
                // 读取像素索引数据
                int pixelLen = frame.Width * frame.Height;
                frame.PixelData = br.ReadBytes(pixelLen);
                Frames.Add(frame);
            }
        }
    }
}