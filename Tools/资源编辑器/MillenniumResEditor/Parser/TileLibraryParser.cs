using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Linq;
using System.Text;
using MillenniumResEditor.Utils;

namespace MillenniumResEditor.Parser
{
    /// <summary>
    /// 千年瓦片图集解析器（.dt / .dh 文件）
    ///
    /// 来自 Delphi 源码 TileCls.pas 的格式规范（32位Delphi编译）：
    ///
    /// .dt 文件结构：
    ///   [0-15]        : 加密文件名前缀(16B) - 跳过
    ///   [16-23]       : 标识 "ATZTIL2"(8B)
    ///   [24-27]       : 瓦片数量 NTile(Int32)
    ///   [28-4123]     : 位置偏移表 FilePos[1024](Int32数组)
    ///   [4124+]       : TTileData[0] + 像素数据[0] + TTileData[1] + 像素数据[1] + ...
    ///
    /// TTileData 结构（sizeof = 116字节，含Bits指针）：
    ///   偏移  大小  字段
    ///   0     4     TileId (Int32)
    ///   4     1     Style (Byte)
    ///   5     3     (填充对齐到4字节边界)
    ///   8     4     nWCell (Int32, 单元格宽=32)
    ///   12    4     nHCell (Int32, 单元格高=24)
    ///   16    4     nBlock (Int32, 帧数)
    ///   20    4     TileWidth (Int32, 像素宽=32)
    ///   24    4     TileHeight (Int32, 像素高=24)
    ///   28   64     MBuffer[64]
    ///   92    4     AniDelay (DWORD) ★ 加载后被覆盖为像素数据偏移!
    ///   96   16     None[4] (Int32*4)
    ///   112   4     Bits (PTAns2Color指针, 文件中存在但忽略)
    ///   总计: 116字节
    ///
    ///   → 后跟 RGB555 像素数据: TileWidth * TileHeight * 2 * nWCell * nHCell * nBlock 字节
    ///
    /// 关键：源码中 pTileData^.AniDelay := pos; // pos是读完TTileData后的文件位置
    ///       即 AniDelay 存储的是该瓦片像素数据的起始偏移量!
    ///
    /// .dh 文件（头索引文件）：
    ///   [0-15]     : 加密文件名前缀(16B) - 跳过
    ///   [16-23]    : 标识 "TILHDF"(8B)
    ///   [24-27]    : 瓦片数量(Int32)
    ///   [28+]      : TTileData[] (仅头部信息，不含像素数据和Bits指针)
    /// </summary>
    public class TileLibraryParser
    {
        // 常量
        public const string TILE_LIB_ID = "ATZTIL2";
        public const string TILE_HEADER_ID = "TILHDF";

        // TFileTileHeader: Ident(8) + NTile(4) + FilePos[1024](4) = 4108字节
        public const int FILE_TILE_HEADER_SIZE = 8 + 4 + 1024 * 4;

        // TTTileData = packed record (Delphi无对齐填充)
        // TileId(4)+Style(1)+nWCell(4)+nHCell(4)+nBlock(4)+
        // TileWidth(4)+TileHeight(4)+MBuffer(64)+AniDelay(4)+None[4](16)+Bits(4)=113B
        public const int TILE_DATA_SIZE = 113;

        /// <summary>图集标识</summary>
        public string LibIdent { get; private set; }

        /// <summary>瓦片总数</summary>
        public int TileCount { get; private set; }

        /// <summary>文件路径</summary>
        public string FilePath { get; private set; }

        /// <summary>是否有对应的 .dh 头文件</summary>
        public bool HasHeaderFile { get; private set; }

        /// <summary>所有瓦片定义</summary>
        public List<TileInfo> Tiles { get; private set; } = new List<TileInfo>();

        /// <summary>原始文件流（用于按需加载像素数据）</summary>
        private FileStream _dataStream;

        /// <summary>加载瓦片图集</summary>
        public void Load(string dtFilePath)
        {
            if (!File.Exists(dtFilePath))
                throw new FileNotFoundException($"瓦片图集文件不存在: {dtFilePath}");

            FilePath = dtFilePath;
            Tiles.Clear();

            // 尝试加载 .dh 头文件
            string dhPath = Path.ChangeExtension(dtFilePath, ".dh");
            HasHeaderFile = false;

            if (File.Exists(dhPath))
            {
                try
                {
                    LoadFromHeaderFile(dhPath);
                    HasHeaderFile = true;
                    return; // 从.dh加载后不需要再解析.dt头
                }
                catch
                {
                    // .dh 加载失败，回退到 .dt
                }
            }

            // 从 .dt 文件加载
            LoadFromFile(dtFilePath);
        }

        /// <summary>从 .dh 头索引文件加载</summary>
        private void LoadFromHeaderFile(string dhPath)
        {
            using var fs = new FileStream(dhPath, FileMode.Open, FileAccess.Read);
            using var br = new BinaryReader(fs);

            // 跳过前16字节加密文件名
            br.ReadBytes(16);

            // 读取头标识
            byte[] identBytes = br.ReadBytes(8);
            string ident = Encoding.ASCII.GetString(identBytes).TrimEnd('\0');

            if (ident != TILE_HEADER_ID)
                throw new InvalidDataException($"无效的瓦片头文件标识: {ident}, 期望: {TILE_HEADER_ID}");

            LibIdent = ident;
            TileCount = br.ReadInt32();

            // 读取所有 TTileData 头部信息
            for (int i = 0; i < TileCount; i++)
            {
                var tile = ReadTileDataHeader(br);
                tile.DataOffset += 16; // 补偿跳过的16字节
                Tiles.Add(tile);
            }
        }

        /// <summary>从 .dt 数据文件加载</summary>
        private void LoadFromFile(string dtPath)
        {
            _dataStream?.Dispose();
            _dataStream = new FileStream(dtPath, FileMode.Open, FileAccess.Read);
            var br = new BinaryReader(_dataStream);

            // 跳过前16字节加密文件名（来自源码: Stream.Seek(16, 0)）
            _dataStream.Seek(16, SeekOrigin.Begin);

            // 读取 TFileTileHeader (4108字节)
            byte[] identBytes = br.ReadBytes(8);
            string ident = Encoding.ASCII.GetString(identBytes).TrimEnd('\0');

            if (ident != TILE_LIB_ID)
                throw new InvalidDataException($"无效的瓦片图集标识: {ident}, 期望: {TILE_LIB_ID}");

            LibIdent = ident;
            TileCount = br.ReadInt32();

            // 读取 FilePos[1024] 偏移表（4096字节）
            int[] filePos = new int[Math.Min(TileCount, 1024)];
            for (int i = 0; i < filePos.Length; i++)
                filePos[i] = br.ReadInt32();

            long dataAreaStart = _dataStream.Position; // 应该是 16 + 4108 = 4124

            // 检测数据区是否为明文（第一个瓦片 TileId 应该是小正整数）
            long probePos = _dataStream.Position;
            if (_dataStream.Length > dataAreaStart + 113)
            {
                int probeTileId = br.ReadInt32();
                _dataStream.Position = probePos;

                // 如果 TileId 不是合理的正整数（1~100000范围），可能是加密/压缩的
                if (probeTileId <= 0 || probeTileId > 100000)
                {
                    throw new InvalidDataException(
                        $"此瓦片图集可能使用了加密或压缩格式（检测到TileId={probeTileId}）。" +
                        Environment.NewLine +
                        "提示：部分 .dt 文件的数据区经过加密，需要配合对应的 .dh 头文件解密。" +
                        Environment.NewLine +
                        "建议：尝试选择其他未加密的 .dt 瓦片图集文件。");
                }
            }

            // 读取所有 TTTileData + 像素数据（来自源码 LoadFromFile）
            for (int i = 0; i < TileCount; i++)
            {
                if (_dataStream.Position + TILE_DATA_SIZE > _dataStream.Length)
                    break;

                // 读取整个 TTTileData 结构体 (packed record, 113字节)
                var tile = ReadTileDataHeader(br);

                // ★ 关键：源码中 pTileData^.AniDelay := pos;
                // pos 是读完 sizeof(TTTileData) 后的文件位置 = 像素数据起始位置!
                tile.DataOffset = _dataStream.Position;

                Tiles.Add(tile);

                // 跳过像素数据到下一个瓦片
                // 源码: Stream.Seek((TW*TH*2) * nWCell*nHCell*nBlock, sofromcurrent)
                if (tile.TileWidth > 0 && tile.TileHeight > 0 &&
                    tile.nWCell > 0 && tile.nHCell > 0 && tile.nBlock > 0)
                {
                    long pixelSize = (long)tile.TileWidth * tile.TileHeight * 2 *
                                      tile.nWCell * tile.nHCell * tile.nBlock;
                    if (_dataStream.Position + pixelSize <= _dataStream.Length)
                        _dataStream.Seek(pixelSize, SeekOrigin.Current);
                    else
                        break; // 文件结束
                }
            }
        }

        /// <summary>读取单个 TTTileData 头部（packed record, 113字节）</summary>
        private TileInfo ReadTileDataHeader(BinaryReader br)
        {
            var tile = new TileInfo();

            tile.TileId = br.ReadInt32();           // +0: 4B
            tile.Style = br.ReadByte();             // +4: 1B (packed, 无填充!)
            tile.nWCell = br.ReadInt32();           // +5: 4B
            tile.nHCell = br.ReadInt32();           // +9: 4B
            tile.nBlock = br.ReadInt32();           // +13: 4B
            tile.TileWidth = br.ReadInt32();        // +17: 4B
            tile.TileHeight = br.ReadInt32();       // +21: 4B
            tile.MBuffer = br.ReadBytes(64);        // +25: 64B
            tile.AniDelay = br.ReadUInt32();        // +89: 4B (加载后被覆盖为数据偏移)
            br.ReadInt32(); br.ReadInt32();         // +93: None[0..1] 8B
            br.ReadInt32(); br.ReadInt32();         // +101: None[2..3] 8B
            br.ReadInt32();                         // +109: Bits指针(4B, 忽略)

            return tile;
        }

        /// <summary>根据 ID 获取瓦片信息</summary>
        public TileInfo GetTileById(int tileId)
        {
            return Tiles.FirstOrDefault(t => t.TileId == tileId);
        }

        /// <summary>获取指定瓦片的图像（RGB555 → Bitmap）</summary>
        /// <param name="tileId">瓦片ID</param>
        /// <param name="frameIndex">动画帧索引（默认0）</param>
        public Bitmap GetTileImage(int tileId, int frameIndex = 0)
        {
            var tile = GetTileById(tileId);
            if (tile == null) return null;

            return GetTileImage(tile, frameIndex);
        }

        /// <summary>获取指定瓦片图像</summary>
        public Bitmap GetTileImage(TileInfo tile, int frameIndex = 0)
        {
            if (tile == null || _dataStream == null) return null;

            int maxFrame = tile.nWCell * tile.nHCell * tile.nBlock;
            if (frameIndex >= maxFrame) frameIndex = 0;

            int w = tile.TileWidth;
            int h = tile.TileHeight;
            int pixelSize = w * h * 2; // RGB555 = 2 bytes/pixel

            lock (_dataStream)
            {
                _dataStream.Seek(tile.DataOffset + pixelSize * frameIndex, SeekOrigin.Begin);
                byte[] rgb555Data = new byte[pixelSize];
                _dataStream.Read(rgb555Data, 0, pixelSize);
                return ConvertRgb555ToBitmap(rgb555Data, w, h);
            }
        }

        /// <summary>从文件直接读取瓦片图像（不依赖缓存流）</summary>
        public Bitmap ReadTileImageDirect(int tileId, int frameIndex = 0)
        {
            var tile = GetTileById(tileId);
            if (tile == null) return null;

            // 边界检查
            if (tile.TileWidth <= 0 || tile.TileHeight <= 0 ||
                tile.TileWidth > 256 || tile.TileHeight > 256)
                return null;

            using var fs = new FileStream(FilePath, FileMode.Open, FileAccess.Read);
            int maxFrame = tile.nWCell * tile.nHCell * tile.nBlock;
            if (maxFrame <= 0 || frameIndex >= maxFrame) frameIndex = 0;

            int w = tile.TileWidth;
            int h = tile.TileHeight;
            int pixelSize = w * h * 2;
            long readPos = tile.DataOffset + pixelSize * frameIndex;

            // 检查偏移是否在文件范围内
            if (readPos < 0 || pixelSize <= 0 || readPos + pixelSize > fs.Length)
                return null;

            fs.Seek(readPos, SeekOrigin.Begin);
            byte[] rgb555Data = new byte[pixelSize];
            int bytesRead = fs.Read(rgb555Data, 0, pixelSize);
            if (bytesRead < pixelSize)
                return null; // 数据不完整

            return ConvertRgb555ToBitmap(rgb555Data, w, h);
        }

        /// <summary>RGB555 转 Bitmap</summary>
        public static Bitmap ConvertRgb555ToBitmap(byte[] rgb555Data, int width, int height)
        {
            if (rgb555Data == null || rgb555Data.Length < width * height * 2)
                return new Bitmap(width, height, PixelFormat.Format32bppArgb);

            var bmp = new Bitmap(width, height, PixelFormat.Format32bppArgb);
            var rect = new Rectangle(0, 0, width, height);
            var bmpData = bmp.LockBits(rect, ImageLockMode.WriteOnly, PixelFormat.Format32bppArgb);

            try
            {
                unsafe
                {
                    byte* dst = (byte*)bmpData.Scan0;
                    int stride = bmpData.Stride;
                    int srcIdx = 0;

                    for (int y = 0; y < height; y++)
                    {
                        byte* row = dst + y * stride;
                        for (int x = 0; x < width; x++)
                        {
                            if (srcIdx + 1 >= rgb555Data.Length) break;

                            // RGB555: RRRRRGGGGGBBBBB (little-endian)
                            ushort pixel = (ushort)(rgb555Data[srcIdx] | (rgb555Data[srcIdx + 1] << 8));
                            srcIdx += 2;

                            // 提取 RGB 各5位并扩展到8位
                            int r = ((pixel >> 10) & 0x1F) << 3;
                            int g = ((pixel >> 5) & 0x1F) << 3;
                            int b = (pixel & 0x1F) << 3;

                            // 补偿低位精度损失
                            r |= r >> 5;
                            g |= g >> 5;
                            b |= b >> 5;

                            int idx = x * 4;
                            row[idx + 0] = (byte)Math.Min(255, b); // B
                            row[idx + 1] = (byte)Math.Min(255, g); // G
                            row[idx + 2] = (byte)Math.Min(255, r); // R
                            row[idx + 3] = 255;                   // A
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

        /// <summary>生成瓦片图集缩略图（用于预览所有瓦片）</summary>
        public Bitmap GenerateThumbnailAtlas(int tilesPerRow = 16, int tileSize = 32)
        {
            if (Tiles.Count == 0) return null;

            int rows = (Tiles.Count + tilesPerRow - 1) / tilesPerRow;
            int imgWidth = tilesPerRow * tileSize;
            int imgHeight = rows * tileSize;

            var bmp = new Bitmap(imgWidth, imgHeight, PixelFormat.Format32bppArgb);
            using var g = Graphics.FromImage(bmp);
            g.Clear(Color.FromArgb(30, 30, 30));

            for (int i = 0; i < Tiles.Count; i++)
            {
                var tile = Tiles[i];
                int col = i % tilesPerRow;
                int row = i / tilesPerRow;

                try
                {
                    using var tileImg = ReadTileImageDirect(tile.TileId);
                    if (tileImg != null)
                    {
                        // 缩放到目标尺寸
                        g.DrawImage(tileImg,
                            col * tileSize, row * tileSize,
                            tileSize, tileSize);
                    }
                }
                catch
                {
                    // 绘制占位符
                    using var brush = new SolidBrush(Color.FromArgb(60, 60 + (i % 3) * 50, 80));
                    g.FillRectangle(brush, col * tileSize, row * tileSize, tileSize - 1, tileSize - 1);
                    g.DrawString($"{tile.TileId}", SystemFonts.DefaultFont, Brushes.White,
                        col * tileSize + 2, row * tileSize + 2);
                }
            }

            return bmp;
        }

        /// <summary>导出单个瓦片为 PNG</summary>
        public void ExportTileToPng(int tileId, string outputPath, int frameIndex = 0)
        {
            var bmp = ReadTileImageDirect(tileId, frameIndex);
            if (bmp != null)
            {
                bmp.Save(outputPath, ImageFormat.Png);
                bmp.Dispose();
            }
        }

        /// <summary>批量导出所有瓦片为 PNG</summary>
        /// <param name="outputDir">输出目录</param>
        /// <returns>成功导出的数量</returns>
        public int ExportAllTiles(string outputDir)
        {
            Directory.CreateDirectory(outputDir);
            int exported = 0;

            foreach (var tile in Tiles)
            {
                try
                {
                    string fileName = Path.Combine(outputDir, $"tile_{tile.TileId:D4}.png");
                    ExportTileToPng(tile.TileId, fileName);
                    exported++;
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"[警告] 导出瓦片 {tile.TileId} 失败: {ex.Message}");
                }
            }

            return exported;
        }

        /// <summary>释放资源</summary>
        public void Dispose()
        {
            _dataStream?.Dispose();
            _dataStream = null;
        }
    }

    /// <summary>
    /// 瓦片信息（对应 Delphi TTileData）
    /// </summary>
    public class TileInfo
    {
        /// <summary>瓦片ID</summary>
        public int TileId { get; set; }

        /// <summary>类型 (0=随机, 1=动画, 2=叠加)</summary>
        public byte Style { get; set; }

        /// <summary>单元格宽度</summary>
        public int nWCell { get; set; } = 32;

        /// <summary>单元格高度</summary>
        public int nHCell { get; set; } = 24;

        /// <summary>块/帧数量</summary>
        public int nBlock { get; set; } = 1;

        /// <summary>瓦片像素宽度</summary>
        public int TileWidth { get; set; } = 32;

        /// <summary>瓦片像素高度</summary>
        public int TileHeight { get; set; } = 24;

        /// <summary>掩码缓冲区</summary>
        public byte[] MBuffer { get; set; }

        /// <summary>像素数据在文件中的偏移位置</summary>
        public uint AniDelay { get; set; }

        /// <summary>数据偏移（运行时计算）</summary>
        public long DataOffset { get; set; }

        public override string ToString()
        {
            return $"Tile[{TileId}] {TileWidth}x{TileHeight} Style={Style} Block={nBlock} Offset={DataOffset}";
        }
    }
}
