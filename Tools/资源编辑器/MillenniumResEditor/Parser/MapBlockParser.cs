using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Linq;
using System.Text;
using MillenniumResEditor.Model;
using MillenniumResEditor.Parser;
using MillenniumResEditor.Utils;

namespace MillenniumResEditor.Parser
{
    /// <summary>
    /// 千年地图块文件解析器（.DT / .dt / .do 等地图资源）
    ///
    /// 格式规范（来自原版 Delphi 源码 MapType.pas + CLMap.pas）：
    ///
    /// 地图由多个 40×40 的块（Block）组成，每个块存储为独立文件：
    ///   - .DT 文件：地面瓦片数据
    ///   .dt 文件：备用地面数据（旧格式）
    ///   .do 文件：对象/装饰物数据
    ///
    /// TMapFileInfo（32字节头）：
    ///   MapIdent[16]     : 标识字符串（如 "MAPDATA1000Y"）
    ///   MapBlockSize(Int32) : 块大小（通常为 40）
    ///   MapWidth(Int32)   : 地图总宽度（格数）
    ///   MapHeight(Int32)  : 地图总高度（格数）
    ///
    /// TMapBlockData（变长）：
    ///   MapBlockIdent[16]      : 块标识
    ///   MapChangedCount(Int32) : 修改计数
    ///   MapBufferArr[1600]     : 40×40 个 TMapCell（每个10字节）
    ///
    /// TMapCell（10字节）：
    ///   TileId(Word)        : 地面瓦片ID (2B)
    ///   TileNumber(Byte)    : 地面瓦片编号 (1B)
    ///   TileOverId(Word)    : 叠加层瓦片ID (2B)
    ///   TileOverNumber(Byte): 叠加层编号 (1B)
    ///   ObjectId(Word)      : 对象ID (2B)
    ///   ObjectNumber(Byte)  : 对象编号 (1B)
    ///   RoofId(Word)        : 屋顶ID (2B)
    ///   boMove(Byte)        : 移动标志 (1B, bit0=可行走,bit1=可飞行)
    ///
    /// 使用场景：
    ///   1. 从 .map 主文件加载地图信息
    ///   2. 按 40×40 分块加载对应的 .DT/.dt/.do 文件
    ///   3. 渲染为完整地图图像或导出为 PNG/TMX 格式
    /// </summary>
    public class MapBlockParser
    {
        // 常量定义（来自 MapType.pas）
        public const int MAP_BLOCK_SIZE = 40;
        public const int CELL_SIZE_BYTES = 10; // TMapCell 大小

        /// <summary>地图标识</summary>
        public string MapIdent { get; private set; }

        /// <summary>块大小（通常40）</summary>
        public int BlockSize { get; private set; } = MAP_BLOCK_SIZE;

        /// <summary>地图总宽度</summary>
        public int MapWidth { get; private set; }

        /// <summary>地图总高度</summary>
        public int MapHeight { get; private set; }

        /// <summary>当前块的 X 索引</summary>
        public int BlockX { get; private set; }

        /// <summary>当前块的 Y 索引</summary>
        public int BlockY { get; private set; }

        /// <summary>修改计数</summary>
        public int ChangedCount { get; private set; }

        /// <summary>瓦片数据数组（40×40 = 1600个）</summary>
        public MapCell[,] Cells { get; private set; }

        /// <summary>原始文件头（32字节）</summary>
        private byte[] _fileInfoHeader;

        /// <summary>块标识</summary>
        private string _blockIdent;

        /// <summary>
        /// 加载地图主文件（.map）获取地图信息
        /// </summary>
        public void LoadMapHeader(string mapFilePath)
        {
            using var fs = new FileStream(mapFilePath, FileMode.Open, FileAccess.Read);
            using var br = new BinaryReader(fs);

            // 读取 TMapFileInfo（32字节）
            _fileInfoHeader = br.ReadBytes(32);

            MapIdent = Encoding.ASCII.GetString(_fileInfoHeader, 0, 16).TrimEnd('\0');
            BlockSize = br.ReadInt32();
            MapWidth = br.ReadInt32();
            MapHeight = br.ReadInt32();
        }

        /// <summary>
        /// 加载单个地图块文件（.DT/.dt/.do）
        /// </summary>
        /// <param name="blockFilePath">块文件路径</param>
        /// <param name="blockX">块X索引</param>
        /// <param name="blockY">块Y索引</param>
        public void LoadBlock(string blockFilePath, int blockX, int blockY)
        {
            BlockX = blockX;
            BlockY = blockY;

            using var fs = new FileStream(blockFilePath, FileMode.Open, FileAccess.Read);
            using var br = new BinaryReader(fs);

            // 读取 TMapBlockData 头部（24字节）
            byte[] identBytes = br.ReadBytes(16);
            _blockIdent = Encoding.ASCII.GetString(identBytes).TrimEnd('\0');
            ChangedCount = br.ReadInt32();

            // 初始化单元格数组
            Cells = new MapCell[BlockSize, BlockSize];

            // 读取所有 TMapCell（40×40 = 1600个，每个10字节）
            for (int y = 0; y < BlockSize; y++)
            {
                for (int x = 0; x < BlockSize; x++)
                {
                    Cells[x, y] = new MapCell
                    {
                        TileId = BinaryUtil.ReadUInt16LE(br),
                        TileNumber = br.ReadByte(),
                        TileOverId = BinaryUtil.ReadUInt16LE(br),
                        TileOverNumber = br.ReadByte(),
                        ObjectId = BinaryUtil.ReadUInt16LE(br),
                        ObjectNumber = br.ReadByte(),
                        RoofId = BinaryUtil.ReadUInt16LE(br),
                        BoMove = br.ReadByte()
                    };
                }
            }
        }

        /// <summary>
        /// 从 PGK 包中加载块数据
        /// </summary>
        public void LoadBlockFromStream(Stream stream, int blockX, int blockY)
        {
            BlockX = blockX;
            BlockY = blockY;

            stream.Position = 0;
            using var br = new BinaryReader(stream);

            byte[] identBytes = br.ReadBytes(16);
            _blockIdent = Encoding.ASCII.GetString(identBytes).TrimEnd('\0');
            ChangedCount = br.ReadInt32();

            Cells = new MapCell[BlockSize, BlockSize];

            for (int y = 0; y < BlockSize; y++)
            {
                for (int x = 0; x < BlockSize; x++)
                {
                    Cells[x, y] = new MapCell
                    {
                        TileId = BinaryUtil.ReadUInt16LE(br),
                        TileNumber = br.ReadByte(),
                        TileOverId = BinaryUtil.ReadUInt16LE(br),
                        TileOverNumber = br.ReadByte(),
                        ObjectId = BinaryUtil.ReadUInt16LE(br),
                        ObjectNumber = br.ReadByte(),
                        RoofId = BinaryUtil.ReadUInt16LE(br),
                        BoMove = br.ReadByte()
                    };
                }
            }
        }

        /// <summary>
        /// 从完整地图数组中提取指定块（用于 .dm 文件加载后预览）
        /// </summary>
        public void LoadFromMapArray(MapCell[,] fullMap, int blockX, int blockY)
        {
            BlockX = blockX;
            BlockY = blockY;

            int mapWidth = fullMap.GetLength(0);
            int mapHeight = fullMap.GetLength(1);

            Cells = new MapCell[BlockSize, BlockSize];

            int startX = blockX * MAP_BLOCK_SIZE;
            int startY = blockY * MAP_BLOCK_SIZE;

            for (int y = 0; y < BlockSize; y++)
            {
                for (int x = 0; x < BlockSize; x++)
                {
                    int gx = startX + x;
                    int gy = startY + y;

                    if (gx < mapWidth && gy < mapHeight)
                        Cells[x, y] = fullMap[gx, gy] ?? new MapCell();
                    else
                        Cells[x, y] = new MapCell();
                }
            }

            _blockIdent = $"BLOCK_{blockX}_{blockY}";
            ChangedCount = 0;
        }

        /// <summary>
        /// 获取指定坐标的单元格
        /// </summary>
        public MapCell GetCell(int x, int y)
        {
            if (x < 0 || x >= BlockSize || y < 0 || y >= BlockSize)
                return null;
            return Cells[x, y];
        }

        /// <summary>
        /// 设置指定坐标的单元格
        /// </summary>
        public void SetCell(int x, int y, MapCell cell)
        {
            if (x < 0 || x >= BlockSize || y < 0 || y >= BlockSize) return;
            Cells[x, y] = cell;
        }

        /// <summary>
        /// 检查是否可以移动到指定位置
        /// </summary>
        public bool IsMovable(int x, int y)
        {
            var cell = GetCell(x, y);
            return cell != null && (cell.BoMove & 0x01) != 0;
        }

        /// <summary>
        /// 检查是否可以飞行到指定位置
        /// </summary>
        public bool IsFlyable(int x, int y)
        {
            var cell = GetCell(x, y);
            return cell != null && (cell.BoMove & 0x02) != 0;
        }

        /// <summary>
        /// 将块保存为文件
        /// </summary>
        public void SaveBlock(string filePath)
        {
            using var fs = new FileStream(filePath, FileMode.Create, FileAccess.Write);
            using var bw = new BinaryWriter(fs);

            // 写入块标识（16字节，不足补零）
            byte[] identBytes = new byte[16];
            string ident = string.IsNullOrEmpty(_blockIdent) ? "MAPBLOCK00" : _blockIdent;
            byte[] tempIdent = Encoding.ASCII.GetBytes(ident);
            Array.Copy(tempIdent, identBytes, Math.Min(tempIdent.Length, 16));
            bw.Write(identBytes);

            bw.Write(ChangedCount);

            // 写入所有单元格
            for (int y = 0; y < BlockSize; y++)
            {
                for (int x = 0; x < BlockSize; x++)
                {
                    var cell = Cells[x, y] ?? new MapCell();
                    BinaryUtil.WriteUInt16LE(bw, cell.TileId);
                    bw.Write(cell.TileNumber);
                    BinaryUtil.WriteUInt16LE(bw, cell.TileOverId);
                    bw.Write(cell.TileOverNumber);
                    BinaryUtil.WriteUInt16LE(bw, cell.ObjectId);
                    bw.Write(cell.ObjectNumber);
                    BinaryUtil.WriteUInt16LE(bw, cell.RoofId);
                    bw.Write(cell.BoMove);
                }
            }
        }

        /// <summary>
        /// 导出块为 PNG 图像（用于可视化调试）
        /// </summary>
        /// <param name="tileWidth">每个瓦片的渲染宽度（像素）</param>
        /// <param name="tileHeight">每个瓦片的渲染高度（像素）</param>
        public Bitmap ExportToImage(int tileWidth = 32, int tileHeight = 24)
        {
            int imgWidth = BlockSize * tileWidth;
            int imgHeight = BlockSize * tileHeight;

            var bmp = new Bitmap(imgWidth, imgHeight, PixelFormat.Format32bppArgb);
            var rect = new Rectangle(0, 0, imgWidth, imgHeight);
            var bmpData = bmp.LockBits(rect, ImageLockMode.WriteOnly, PixelFormat.Format32bppArgb);

            try
            {
                unsafe
                {
                    byte* dst = (byte*)bmpData.Scan0;
                    int stride = bmpData.Stride;

                    for (int y = 0; y < BlockSize; y++)
                    {
                        for (int x = 0; x < BlockSize; x++)
                        {
                            var cell = Cells[x, y];

                            // 根据瓦片 ID 生成伪彩色（实际应使用瓦片集渲染）
                            Color color = GenerateTileColor(cell.TileId, cell.TileOverId);

                            int px = x * tileWidth;
                            int py = y * tileHeight;

                            // 填充该瓦片区域
                            for (int ty = 0; ty < tileHeight; ty++)
                            {
                                byte* row = dst + (py + ty) * stride;
                                for (int tx = 0; tx < tileWidth; tx++)
                                {
                                    int idx = (px + tx) * 4;
                                    row[idx + 0] = color.B;
                                    row[idx + 1] = color.G;
                                    row[idx + 2] = color.R;
                                    row[idx + 3] = color.A;
                                }
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

        /// <summary>
        /// 根据瓦片 ID 生成伪彩色（用于可视化）
        /// 实际应用中应该从瓦片集（TIL/PAL）加载真实图像
        /// </summary>
        private static Color GenerateTileColor(ushort tileId, ushort overId)
        {
            if (tileId == 0 && overId == 0)
                return Color.FromArgb(0, 0, 0, 0); // 完全透明

            // 使用 ID 生成确定性颜色
            int r = (tileId * 37 + 100) % 256;
            int g = (tileId * 51 + 80) % 256;
            int b = (tileId * 73 + 150) % 256;

            // 如果有叠加层，稍微提亮
            if (overId > 0)
            {
                r = Math.Min(255, r + 30);
                g = Math.Min(255, g + 30);
                b = Math.Min(255, b + 30);
            }

            return Color.FromArgb(255, r, g, b);
        }

        /// <summary>
        /// 导出为 TMX 格式（Tiled Map Editor 兼容）
        /// 用于在其他地图编辑器中编辑
        /// </summary>
        public string ExportToTmx()
        {
            var sb = new StringBuilder();
            sb.AppendLine("<?xml version=\"1.0\" encoding=\"UTF-8\"?>");
            sb.AppendLine("<map version=\"1.5\" tiledversion=\"1.7.2\" " +
                         $"orientation=\"orthogonal\" renderorder=\"right-down\" " +
                         $"width=\"{BlockSize}\" height=\"{BlockSize}\" " +
                         $"tilewidth=\"32\" tileheight=\"24\" infinite=\"0\">");

            // 地面层
            sb.AppendLine(" <layer id=\"1\" name=\"Ground\" width=\"40\" height=\"40\">");
            sb.AppendLine("  <data encoding=\"csv\">");
            for (int y = 0; y < BlockSize; y++)
            {
                var line = new List<string>();
                for (int x = 0; x < BlockSize; x++)
                {
                    var cell = Cells[x, y];
                    line.Add(cell.TileId.ToString());
                }
                sb.AppendLine($"   {string.Join(",", line)},");
            }
            sb.AppendLine("  </data>");
            sb.AppendLine(" </layer>");

            // 叠加层
            sb.AppendLine(" <layer id=\"2\" name=\"Overlay\" width=\"40\" height=\"40\">");
            sb.AppendLine("  <data encoding=\"csv\">");
            for (int y = 0; y < BlockSize; y++)
            {
                var line = new List<string>();
                for (int x = 0; x < BlockSize; x++)
                {
                    var cell = Cells[x, y];
                    line.Add(cell.TileOverId.ToString());
                }
                sb.AppendLine($"   {string.Join(",", line)},");
            }
            sb.AppendLine("  </data>");
            sb.AppendLine(" </layer>");

            // 对象层
            sb.AppendLine(" <objectgroup id=\"3\" name=\"Objects\">");
            for (int y = 0; y < BlockSize; y++)
            {
                for (int x = 0; x < BlockSize; x++)
                {
                    var cell = Cells[x, y];
                    if (cell.ObjectId > 0)
                    {
                        sb.AppendLine($"  <object id=\"{cell.ObjectId}\" x=\"{x * 32}\" y=\"{y * 24}\" " +
                                     $"width=\"32\" height=\"24\" type=\"obj_{cell.ObjectId}\"/>");
                    }
                }
            }
            sb.AppendLine(" </objectgroup>");

            sb.AppendLine("</map>");
            return sb.ToString();
        }
    }

    /// <summary>
    /// 地图单元格（对应 Delphi TMapCell）
    /// </summary>
    public class MapCell
    {
        /// <summary>地面瓦片ID</summary>
        public ushort TileId { get; set; }

        /// <summary>地面瓦片编号（同一瓦片集的不同帧）</summary>
        public byte TileNumber { get; set; }

        /// <summary>叠加层瓦片ID</summary>
        public ushort TileOverId { get; set; }

        /// <summary>叠加层瓦片编号</summary>
        public byte TileOverNumber { get; set; }

        /// <summary>对象ID</summary>
        public ushort ObjectId { get; set; }

        /// <summary>对象编号</summary>
        public byte ObjectNumber { get; set; }

        /// <summary>屋顶ID</summary>
        public ushort RoofId { get; set; }

        /// <summary>移动标志（bit0=可行走, bit1=可飞行）</summary>
        public byte BoMove { get; set; }

        public override string ToString()
        {
            return $"Tile={TileId}:{TileNumber} Over={TileOverId}:{TileOverNumber} " +
                   $"Obj={ObjectId}:{ObjectNumber} Roof={RoofId} Move={BoMove:X2}";
        }
    }
}
