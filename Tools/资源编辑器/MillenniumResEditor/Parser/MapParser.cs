using System;
using System.IO;
using System.Linq;
using MillenniumResEditor.Model;
using MillenniumResEditor.Utils;

namespace MillenniumResEditor.Parser
{
    /// <summary>
    /// 千年 .map 二进制地图解析器（读+写）
    /// 格式：128B头 + 宽(2) + 高(2) + 每个瓦片3B(Low/High/Attr)
    /// </summary>
    public class MapParser
    {
        public ushort Width { get; private set; }
        public ushort Height { get; private set; }
        public MapTile[] Tiles { get; private set; }
        private byte[] _fileHeader;

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

            // 3. 循环读取所有瓦片
            for (int i = 0; i < total; i++)
            {
                Tiles[i] = new MapTile
                {
                    Low = br.ReadByte(),
                    High = br.ReadByte(),
                    Attr = br.ReadByte()
                };
            }
        }

        /// <summary>保存为原版 .map 文件</summary>
        public void Save(string filePath)
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
                bw.Write(tile.Low);
                bw.Write(tile.High);
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
    }
}