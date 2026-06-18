using System;
using System.IO;
using System.Linq;
using MillenniumResEditor.Model;
using MillenniumResEditor.Utils;

namespace MillenniumResEditor.Parser
{
    /// <summary>
    /// 千年 .map 二进制地图解析器（读+写）
    /// 支持两种格式：
    ///   旧格式：128B头 + 宽(2) + 高(2) + 每个瓦片3B(Low/High/Attr) [8位ID]
    ///   新格式：128B头 + 宽(2) + 高(2) + 每个瓦片5B(Low(2)/High(2)/Attr) [16位ID]
    /// </summary>
    public class MapParser
    {
        public ushort Width { get; private set; }
        public ushort Height { get; private set; }
        public MapTile[] Tiles { get; private set; }
        private byte[] _fileHeader;
        private bool _isNewFormat; // 是否为新格式(16位ID)

        /// <summary>加载 .map 文件</summary>
        public void Load(string filePath)
        {
            using var fs = new FileStream(filePath, FileMode.Open, FileAccess.Read);
            using var br = new BinaryReader(fs);

            // 1. 读取128字节文件头（原样保留，不修改）
            _fileHeader = br.ReadBytes(128);

            // 2. 读取宽高（小端）
            Width = BinaryUtil.ReadUInt16LE(br);
            Height = BinaryUtil.ReadUInt16LE(br);

            int total = Width * Height;
            Tiles = new MapTile[total];

            // 3. 自动检测文件格式：通过计算期望文件大小来判断
            // 新格式: 每瓦片5字节(low:2 + high:2 + attr:1)
            // 旧格式: 每瓦片3字节(low:1 + high:1 + attr:1)
            long expectedSizeNewFormat = 128 + 4 + (long)total * 5;
            long expectedSizeOldFormat = 128 + 4 + (long)total * 3;
            _isNewFormat = Math.Abs(fs.Length - expectedSizeNewFormat) <= Math.Abs(fs.Length - expectedSizeOldFormat);

            // 4. 循环读取所有瓦片
            for (int i = 0; i < total; i++)
            {
                Tiles[i] = new MapTile();
                if (_isNewFormat)
                {
                    // 新格式：16位瓦片ID
                    Tiles[i].Low = BinaryUtil.ReadUInt16LE(br);
                    Tiles[i].High = BinaryUtil.ReadUInt16LE(br);
                }
                else
                {
                    // 旧格式：8位瓦片ID
                    Tiles[i].Low = br.ReadByte();
                    Tiles[i].High = br.ReadByte();
                }
                Tiles[i].Attr = br.ReadByte();
            }
        }

        /// <summary>保存为 .map 文件（默认使用新格式）</summary>
        public void Save(string filePath)
        {
            Save(filePath, true);
        }

        /// <summary>保存为 .map 文件</summary>
        /// <param name="useNewFormat">是否使用新格式(16位ID)</param>
        public void Save(string filePath, bool useNewFormat)
        {
            using var fs = new FileStream(filePath, FileMode.Create, FileAccess.Write);
            using var bw = new BinaryWriter(fs);

            // 1. 写入原文件头
            bw.Write(_fileHeader);

            // 2. 写入宽高
            BinaryUtil.WriteUInt16LE(bw, Width);
            BinaryUtil.WriteUInt16LE(bw, Height);

            // 3. 写入所有瓦片
            foreach (var tile in Tiles)
            {
                if (useNewFormat)
                {
                    // 新格式：16位瓦片ID
                    BinaryUtil.WriteUInt16LE(bw, tile.Low);
                    BinaryUtil.WriteUInt16LE(bw, tile.High);
                }
                else
                {
                    // 旧格式：8位瓦片ID（高位会被截断）
                    bw.Write((byte)tile.Low);
                    bw.Write((byte)tile.High);
                }
                bw.Write(tile.Attr);
            }
        }

        /// <summary>根据坐标获取瓦片</summary>
        public MapTile GetTile(int x, int y)
        {
            int idx = y * Width + x;
            if (idx < 0 || idx >= Tiles.Length) return null;
            return Tiles[idx];
        }

        /// <summary>修改瓦片</summary>
        public void SetTile(int x, int y, MapTile tile)
        {
            int idx = y * Width + x;
            if (idx < 0 || idx >= Tiles.Length) return;
            Tiles[idx] = tile;
        }

        /// <summary>获取当前文件格式</summary>
        public bool IsNewFormat => _isNewFormat;
    }
}